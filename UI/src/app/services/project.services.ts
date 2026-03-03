import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { IRequest } from '../models/request.model';

@Injectable({
  providedIn: 'root'
})
export class ProjectServices {
  private urlPath: string = 'http://localhost:8080'; // Временное решение

  constructor(private http: HttpClient) {
  }

  getAll(page: number = 1, searchName: string = ''): Observable<IRequest> {
    const params = new HttpParams()
      .set('page', String(page))
      .set('limit', '20')
      .set('search', searchName || '');

    const url = `${this.urlPath}/projects`;
    return this.http.get<IRequest>(url, { params }).pipe(
      catchError(err => {
        console.error('ProjectServices.getAll error', err);
        return throwError(() => err);
      })
    );
  }

  addProject(key: string): Observable<IRequest> {
    const url = `${this.urlPath}/connector/updateProject`;
    return this.http.post<IRequest>(url, { key }).pipe(
      catchError(err => {
        console.error('ProjectServices.addProject error', err);
        return throwError(() => err);
      })
    );
  }

  deleteProject(id: number): Observable<IRequest> {
    const url = `${this.urlPath}/projects/${id}`;
    return this.http.delete<IRequest>(url).pipe(
      catchError(err => {
        console.error('ProjectServices.deleteProject error', err);
        return throwError(() => err);
      })
    );
  }
}
