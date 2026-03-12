package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	filePrefix    = "wayback-"
	fileSuffix    = ".log"
	dateFormat    = "2006-01-02"
	retentionDays = 7
)

// Logger manages log file rotation and cleanup.
type Logger struct {
	dir     string
	mu      sync.Mutex
	file    *os.File
	curDate string
	stopCh  chan struct{}
}

// Setup initializes the logging system: creates log dir, opens today's
// log file, redirects stdlib log + gin output, and starts the cleanup goroutine.
func Setup(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	l := &Logger{dir: logDir, stopCh: make(chan struct{})}
	if err := l.rotate(); err != nil {
		return nil, err
	}

	// Clean old logs on startup
	l.cleanup()

	// Background: rotate at midnight + periodic cleanup
	go l.backgroundLoop()

	return l, nil
}

// Close stops the background goroutine and closes the current log file.
func (l *Logger) Close() {
	close(l.stopCh)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
	}
}

// Dir returns the log directory path.
func (l *Logger) Dir() string {
	return l.dir
}

// rotate opens (or switches to) today's log file.
func (l *Logger) rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().Format(dateFormat)
	if today == l.curDate && l.file != nil {
		return nil
	}

	if l.file != nil {
		l.file.Close()
	}

	filename := filepath.Join(l.dir, filePrefix+today+fileSuffix)
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	l.file = f
	l.curDate = today

	// Redirect stdlib log and gin to both stdout and file
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	gin.DefaultWriter = mw
	gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, f)

	return nil
}

// cleanup removes log files older than retentionDays.
func (l *Logger) cleanup() {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(l.dir)
	if err != nil {
		log.Printf("[logging] failed to read log dir: %v", err)
		return
	}

	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, filePrefix) || !strings.HasSuffix(name, fileSuffix) {
			continue
		}
		dateStr := strings.TrimPrefix(name, filePrefix)
		dateStr = strings.TrimSuffix(dateStr, fileSuffix)
		t, err := time.Parse(dateFormat, dateStr)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			path := filepath.Join(l.dir, name)
			if err := os.Remove(path); err != nil {
				log.Printf("[logging] failed to remove old log %s: %v", name, err)
			} else {
				log.Printf("[logging] removed old log: %s", name)
			}
		}
	}
}

// backgroundLoop handles daily rotation and cleanup.
func (l *Logger) backgroundLoop() {
	for {
		now := time.Now()
		// Next midnight
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 1, 0, now.Location())
		timer := time.NewTimer(next.Sub(now))

		select {
		case <-timer.C:
			if err := l.rotate(); err != nil {
				log.Printf("[logging] rotation failed: %v", err)
			}
			l.cleanup()
		case <-l.stopCh:
			timer.Stop()
			return
		}
	}
}

// ListLogFiles returns metadata about available log files, sorted newest first.
type LogFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Date string `json:"date"`
}

func (l *Logger) ListLogFiles() ([]LogFileInfo, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, err
	}

	var files []LogFileInfo
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, filePrefix) || !strings.HasSuffix(name, fileSuffix) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		dateStr := strings.TrimPrefix(name, filePrefix)
		dateStr = strings.TrimSuffix(dateStr, fileSuffix)
		files = append(files, LogFileInfo{
			Name: name,
			Size: info.Size(),
			Date: dateStr,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Date > files[j].Date
	})
	return files, nil
}

// ReadLogFile reads the last N lines of a log file.
func (l *Logger) ReadLogFile(filename string, tail int) (string, error) {
	// Sanitize filename to prevent path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
		return "", fmt.Errorf("invalid filename")
	}
	if !strings.HasPrefix(filename, filePrefix) || !strings.HasSuffix(filename, fileSuffix) {
		return "", fmt.Errorf("invalid log filename")
	}

	path := filepath.Join(l.dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)
	if tail <= 0 {
		tail = 500
	}

	lines := strings.Split(content, "\n")
	if len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	return strings.Join(lines, "\n"), nil
}
