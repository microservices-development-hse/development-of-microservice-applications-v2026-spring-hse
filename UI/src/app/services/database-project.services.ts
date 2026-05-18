import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { IRequest } from '../models/request.model';
import { IRequestObject } from '../models/requestObj.model';

@Injectable({
  providedIn: 'root'
})
export class DatabaseProjectServices {
  private urlPath: string = 'http://localhost:8000/api/v1';

  constructor(private http: HttpClient) {}

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

  recalculateProject(id: string): Observable<any> {
    const url = `${this.urlPath}/projects/${encodeURIComponent(id)}/analytics/recalculate`;
    return this.http.post<any>(url, {}).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.recalculateProject error', err);
        return throwError(() => err);
      })
    );
  }

  getAnalytics(id: string, type: string): Observable<any> {
    const url = `${this.urlPath}/projects/${encodeURIComponent(id)}/analytics`;
    const params = new HttpParams().set('type', type);

    return this.http.get<any>(url, { params }).pipe(
      catchError(err => {
        console.error('DatabaseProjectServices.getAnalytics error', err);
        return throwError(() => err);
      })
    );
  }
}
