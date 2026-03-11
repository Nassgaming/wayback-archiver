// TypeScript interfaces for the Wayback archiver

export interface CaptureData {
  url: string;
  title: string;
  html: string;
  headers?: Record<string, string>;
}

export interface ArchiveResponse {
  status: string;
  page_id: number;
  action: 'created' | 'unchanged' | 'updated';
}
