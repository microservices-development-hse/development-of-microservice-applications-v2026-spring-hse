export interface Project {
  key: string;
  name: string;
  totalIssues?: number;
  openIssues?: number;
  createdAt?: string;
  analyzed?: boolean;
}
