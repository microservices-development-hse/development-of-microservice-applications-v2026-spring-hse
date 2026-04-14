import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { map } from 'rxjs/operators';
import { IRequest } from '../models/request.model';
import { IRequestObject } from '../models/requestObj.model';

@Injectable({
  providedIn: 'root'
})
export class DatabaseProjectServices {
  private urlPath: string = 'http://localhost:8000/api/v1'; // Временное решение

  constructor(private http: HttpClient) {
  }

   getAll(): Observable<IRequest> {
    const url = `${this.urlPath}/projects`;

    return this.http.get<any>(url).pipe(
      map((response: any) => {
        const projects = (response.projects || []).map((p: any) => ({
          Existence: true,
          Id: p.id,
          Key: p.key,
          Name: p.title,
          Url: p.url
        }));

        return {
          projects,
          pageInfo: {
            currentPage: response.pageInfo?.currentPage ?? 1,
            projectsCount: response.pageInfo?.projectsCount ?? projects.length,
            pageCount: response.pageInfo?.pageCount ?? 1
          }
        };
      }),
      catchError(err => {
        console.error('DatabaseProjectServices.getAll error', err);
        return throwError(() => err);
      })
    );
  }

  getProjectStatByID(id: string): Observable<any> {
    const url = `${this.urlPath}/projects/${encodeURIComponent(id)}`;
    return this.http.get<any>(url).pipe(
      map((response: any) => ({
        data: {
          allIssuesCount: response.stats?.total_tasks ?? 0,
          openIssuesCount: response.stats?.open_tasks ?? 0,
          closeIssuesCount: response.stats?.closed_tasks ?? 0,
          reopenedIssuesCount: response.stats?.reopened_tasks ?? 0,
          resolvedIssuesCount: response.stats?.resolved_tasks ?? 0,
          progressIssuesCount: response.stats?.in_progress_tasks ?? 0,
          averageTime: response.stats?.avg_lead_time_h ?? 0,
          averageIssuesCount: response.stats?.avg_daily_weekly ?? 0
        }
      })),
      catchError(err => {
        console.error('DatabaseProjectServices.getProjectStatByID error', err);
        return throwError(() => err);
      })
    );
  }

  getComplitedGraph(taskNumber: string, projectName: Array<string>): Observable<IRequestObject> {
    const url = `${this.urlPath}/graph/compare`;
    return this.http.post<IRequestObject>(url, { task: taskNumber, projects: projectName }).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.getComplitedGraph error', err);
        return throwError(() => err);
      })
    );
  }

  getGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    const params = new HttpParams().set('task', taskNumber).set('project', projectName);
    const url = `${this.urlPath}/graph/get`;
    return this.http.get<IRequestObject>(url, { params }).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.getGraph error', err);
        return throwError(() => err);
      })
    );
  }

  makeGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    const url = `${this.urlPath}/graph/make`;
    return this.http.post<IRequestObject>(url, { task: taskNumber, project: projectName }).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.makeGraph error', err);
        return throwError(() => err);
      })
    );
  }

  deleteGraphs(projectName: string): Observable<IRequestObject> {
    const url = `${this.urlPath}/graph/delete`;
    return this.http.request<IRequestObject>('delete', url, { body: { project: projectName } }).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.deleteGraphs error', err);
        return throwError(() => err);
      })
    );
  }

  isAnalyzed(projectName: string): Observable<IRequestObject> {
    const url = `${this.urlPath}/isAnalyzed`;
    const params = new HttpParams().set('project', projectName);
    return this.http.get<IRequestObject>(url, { params }).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.isAnalyzed error', err);
        return throwError(() => err);
      })
    );
  }

  isEmpty(projectName: string): Observable<IRequestObject> {
    const url = `${this.urlPath}/isEmpty`;
    const params = new HttpParams().set('project', projectName);
    return this.http.get<IRequestObject>(url, { params }).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.isEmpty error', err);
        return throwError(() => err);
      })
    );
  }
}
