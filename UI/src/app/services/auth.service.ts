import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { ConfigurationService } from './configuration.services';
import { IAuthSession, ILoginResponse, IRegisterResponse } from '../models/auth.model';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly storageKey = 'jira-analyzer-session';

  constructor(
    private http: HttpClient,
    private configurationService: ConfigurationService
  ) {}

  private get authUrl(): string {
    return this.configurationService.getValue('authUrl', 'http://localhost:8083');
  }

  login(email: string, password: string): Observable<ILoginResponse> {
    return this.http.post<ILoginResponse>(`${this.authUrl}/login`, { email, password }).pipe(
      tap((response) => this.saveSession(response))
    );
  }

  register(email: string, password: string): Observable<IRegisterResponse> {
    return this.http.post<IRegisterResponse>(`${this.authUrl}/register`, { email, password });
  }

  logout(): void {
    localStorage.removeItem(this.storageKey);
  }

  clearSession(): void {
    localStorage.removeItem(this.storageKey);
  }

  isAuthenticated(): boolean {
    const session = this.getSession();
    if (!session) {
      return false;
    }

    if (session.expiresAt <= Date.now()) {
      this.logout();
      return false;
    }

    return true;
  }

  getToken(): string | null {
    return this.getSession()?.token ?? null;
  }

  getEmail(): string | null {
    return this.getSession()?.email ?? null;
  }

  private saveSession(response: ILoginResponse): void {
    const expiresAtMs = response.expiresAt > 1_000_000_000_000
      ? response.expiresAt
      : response.expiresAt * 1000;

    const session: IAuthSession = {
      token: response.token,
      email: response.email,
      expiresAt: expiresAtMs
    };

    localStorage.setItem(this.storageKey, JSON.stringify(session));
  }

  private getSession(): IAuthSession | null {
    const raw = localStorage.getItem(this.storageKey);
    if (!raw) {
      return null;
    }

    try {
      return JSON.parse(raw) as IAuthSession;
    } catch {
      return null;
    }
  }
}

