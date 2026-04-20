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
    const url = `${this.urlPath}/connector/projects`;
    const search = searchName.trim().toLowerCase();

    return this.http.get<any>(url).pipe(
      map((response: any) => {
        const list = Array.isArray(response) ? response : (response.projects || []);

        const filtered = list.filter((p: any) => {
          if (!search) {
            return true;
          }

          const key = (p.key || '').toLowerCase();
          const name = (p.name || p.title || '').toLowerCase();
          const url = (p.url || '').toLowerCase();

          return key.includes(search) || name.includes(search) || url.includes(search);
        });

        const projects = filtered.map((p: any) => ({
          Existence: false,
          Id: 0,
          Key: p.key,
          Name: p.name || p.key,
          Url: p.url || ''
        }));
  
        return {
          projects,
          pageInfo: {
            currentPage: page,
            projectsCount: projects.length,
            pageCount: Math.max(1, Math.ceil(projects.length / 10))
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
