import { Injectable } from '@angular/core';
import {
  HttpErrorResponse,
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { AuthService } from './auth.service';
import { ConfigurationService } from './configuration.services';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(
    private authService: AuthService,
    private configurationService: ConfigurationService,
    private router: Router
  ) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const authUrl = this.configurationService.getValue('authUrl', 'http://localhost:8083');
    const backendUrl = this.configurationService.getValue('pathUrl', 'http://localhost:8000');

    const skipAuth =
      req.url.startsWith(authUrl) ||
      req.url.includes('/assets/') ||
      req.url.includes('/login');

    let authReq = req;

    if (!skipAuth && req.url.startsWith(backendUrl)) {
      const token = this.authService.getToken();
      if (token) {
        authReq = req.clone({
          setHeaders: {
            Authorization: `Bearer ${token}`
          }
        });
      }
    }

    return next.handle(authReq).pipe(
      catchError((error: HttpErrorResponse) => {
        if (error.status === 401 && !skipAuth) {
          this.authService.clearSession();
          if (this.router.url !== '/login') {
            this.router.navigate(['/login'], {
              queryParams: { returnUrl: this.router.url }
            });
          }
        }
        return throwError(() => error);
      })
    );
  }
}

