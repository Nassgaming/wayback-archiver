package api

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"wayback/internal/models"
)

// rewriteResourcePathsWithResources 将资源路径重写为代理格式
func rewriteResourcePathsWithResources(html string, pageID int64, timestamp string, resources []models.Resource, pageURL string) string {
	result := html

	// 首先处理 /archive/resources/ 格式的路径
	result = archivePathRe.ReplaceAllString(result, fmt.Sprintf("/archive/%d/%smp_/resources/", pageID, timestamp))

	// 然后处理每个资源的 URL
	for _, resource := range resources {
		// 构建代理 URL
		proxyURL := fmt.Sprintf("/archive/%d/%smp_/%s", pageID, timestamp, resource.URL)

		// 转义原始 URL 用于正则表达式
		escapedURL := regexp.QuoteMeta(resource.URL)

		// 替换绝对 URL
		patterns := []string{
			`src=["']` + escapedURL + `["']`,
			`href=["']` + escapedURL + `["']`,
			`url\(["']?` + escapedURL + `["']?\)`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			result = re.ReplaceAllStringFunc(result, func(match string) string {
				if strings.Contains(match, `src=`) {
					return `src="` + proxyURL + `"`
				} else if strings.Contains(match, `href=`) {
					return `href="` + proxyURL + `"`
				} else if strings.Contains(match, `url(`) {
					return `url("` + proxyURL + `")`
				}
				return match
			})
		}

		// 处理协议相对 URL（如 //example.com/path）
		// 从完整 URL 中移除协议部分
		protocolRelativeURL := strings.TrimPrefix(resource.URL, "https:")
		protocolRelativeURL = strings.TrimPrefix(protocolRelativeURL, "http:")

		if protocolRelativeURL != resource.URL && strings.HasPrefix(protocolRelativeURL, "//") {
			escapedProtocolRelativeURL := regexp.QuoteMeta(protocolRelativeURL)
			protocolRelativePatterns := []string{
				`src=["']` + escapedProtocolRelativeURL + `["']`,
				`href=["']` + escapedProtocolRelativeURL + `["']`,
				`url\(["']?` + escapedProtocolRelativeURL + `["']?\)`,
			}

			for _, pattern := range protocolRelativePatterns {
				re := regexp.MustCompile(pattern)
				result = re.ReplaceAllStringFunc(result, func(match string) string {
					if strings.Contains(match, `src=`) {
						return `src="` + proxyURL + `"`
					} else if strings.Contains(match, `href=`) {
						return `href="` + proxyURL + `"`
					} else if strings.Contains(match, `url(`) {
						return `url("` + proxyURL + `")`
					}
					return match
				})
			}
		}

		// 处理相对路径
		// 从资源 URL 中提取文件名
		parsedURL, err := url.Parse(resource.URL)
		if err != nil {
			continue
		}
		filename := path.Base(parsedURL.Path)

		// 替换 ./filename 和 filename 格式
		relativePatterns := []string{
			`src=["']\.?/?` + regexp.QuoteMeta(filename) + `["']`,
			`href=["']\.?/?` + regexp.QuoteMeta(filename) + `["']`,
		}

		for _, pattern := range relativePatterns {
			re := regexp.MustCompile(pattern)
			result = re.ReplaceAllStringFunc(result, func(match string) string {
				if strings.Contains(match, `src=`) {
					return `src="` + proxyURL + `"`
				} else if strings.Contains(match, `href=`) {
					return `href="` + proxyURL + `"`
				}
				return match
			})
		}

		// 处理以 / 开头的绝对路径（如 /assets/style.css）
		// 提取资源URL的路径部分
		resourcePath := parsedURL.Path
		if resourcePath != "" {
			// 构建带查询参数的完整路径（如果有的话）
			pathWithQuery := resourcePath
			if parsedURL.RawQuery != "" {
				pathWithQuery = resourcePath + "?" + parsedURL.RawQuery
			}

			escapedPath := regexp.QuoteMeta(pathWithQuery)
			absolutePathPatterns := []string{
				`src=["']` + escapedPath + `["']`,
				`href=["']` + escapedPath + `["']`,
				`url\(["']?` + escapedPath + `["']?\)`,
				// 处理HTML实体编码的引号：url(&quot;/path&quot;)
				`url\(&quot;` + escapedPath + `&quot;\)`,
			}

			for _, pattern := range absolutePathPatterns {
				re := regexp.MustCompile(pattern)
				result = re.ReplaceAllStringFunc(result, func(match string) string {
					if strings.Contains(match, `src=`) {
						return `src="` + proxyURL + `"`
					} else if strings.Contains(match, `href=`) {
						return `href="` + proxyURL + `"`
					} else if strings.Contains(match, `url(`) {
						return `url("` + proxyURL + `")`
					}
					return match
				})
			}
		}
	}

	return result
}
