import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class ConfigService {
  private cfg: any = null;

  constructor(private http: HttpClient) {}

  load(): Promise<void> {
    // path: /config.json (public/) или /assets/config.json
    return firstValueFrom(this.http.get('/config.json'))
      .then(cfg => { this.cfg = cfg; })
      .catch(() => { this.cfg = {}; }); // fallback
  }

  get apiBaseUrl(): string {
    return this.cfg?.apiBaseUrl || 'http://localhost:8080/api/v1';
  }
}
