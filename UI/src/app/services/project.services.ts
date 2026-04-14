import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { IRequest } from '../models/request.model';
import { IProj } from '../models/proj.model';

@Injectable({
  providedIn: 'root'
})
export class ProjectServices {
  private urlPath: string = 'http://localhost:8000/api/v1';

  constructor(private http: HttpClient) {}

  getAll(page: number = 1, searchName: string = ''): Observable<IRequest> {
    const params = new HttpParams()
      .set('page', String(page))
      .set('limit', '20')
      .set('search', searchName || '');

    const url = `${this.urlPath}/connector/projects`;

    return this.http.get<any>(url, { params }).pipe(
      map((response: any) => {
        const list = Array.isArray(response) ? response : (response.projects || []);
  
        const projects = list.map((p: any) => ({
          Existence: false,
          Id: 0,
          Key: p.key,
          Name: p.name || p.key,
          Url: p.url || ''
        }));

        return {
          projects,
          pageInfo: {
            currentPage: 1,
            projectsCount: projects.length,
            pageCount: 1
          }
        };
      }),
      catchError(err => {
        console.error('ProjectServices.getAll error', err);
        return throwError(() => err);
      })
    );
  }

  addProject(project: IProj): Observable<any> {
    const url = `${this.urlPath}/connector/import`;
    return this.http.post<any>(url, {
      project_key: project.Key
    }).pipe(
      catchError(err => {
        console.error('ProjectServices.addProject error', err);
        return throwError(() => err);
      })
    );
  }

  deleteProject(id: number): Observable<any> {
    const url = `${this.urlPath}/projects/${id}`;
    return this.http.delete<any>(url).pipe(
      catchError(err => {
        console.error('ProjectServices.deleteProject error', err);
        return throwError(() => err);
      })
    );
  }
}
