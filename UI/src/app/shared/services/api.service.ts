import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Project } from '../models/project.model';
import { ConfigService } from './config.service';

@Injectable({ providedIn: 'root' })
export class ApiService {
  constructor(private http: HttpClient, private cfg: ConfigService) {}

  private base(path: string) {
    // cfg.apiBaseUrl уже загружен благодаря APP_INITIALIZER
    return `${this.cfg.apiBaseUrl}${path}`;
  }

  getProjects(limit = 20, page = 1, search = ''): Observable<{ data: Project[]; total: number; }> {
    const q = `?limit=${limit}&page=${page}&search=${encodeURIComponent(search)}`;
    return this.http.get<{ data: Project[]; total: number; }>(this.base(`/projects${q}`));
  }

  addProject(projectKey: string): Observable<any> {
    return this.http.post(this.base('/connector/updateProject'), { key: projectKey });
  }

  isAnalyzed(projectKey: string): Observable<{ analyzed: boolean }> {
    return this.http.get<{ analyzed: boolean }>(this.base(`/isAnalyzed?project=${encodeURIComponent(projectKey)}`));
  }

  makeGraph(taskNumber: number, projectKey: string): Observable<any> {
    return this.http.post(this.base('/graph/make'), { task: taskNumber, project: projectKey });
  }

  getGraph(taskNumber: number, projectKey: string): Observable<any> {
    return this.http.get(this.base(`/graph/get?task=${taskNumber}&project=${encodeURIComponent(projectKey)}`));
  }

  compareGraphs(taskNumber: number, projectKeys: string[]): Observable<any> {
    return this.http.post(this.base('/graph/compare'), { task: taskNumber, projects: projectKeys });
  }
}
