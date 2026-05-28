import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-login-page',
  templateUrl: './login-page.component.html',
  styleUrls: ['./login-page.component.css']
})
export class LoginPageComponent implements OnInit {
  mode: 'login' | 'register' = 'login';
  email = '';
  password = '';
  loading = false;
  error = '';
  success = '';

  constructor(
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit(): void {
    if (this.authService.isAuthenticated()) {
      this.router.navigate(['/projects']);
    }
  }

  toggleMode(): void {
    this.mode = this.mode === 'login' ? 'register' : 'login';
    this.error = '';
    this.success = '';
    this.loading = false;
  }

  onSubmit(): void {
    this.error = '';
    this.success = '';
    this.loading = true;

    if (this.mode === 'login') {
      this.authService.login(this.email, this.password).subscribe({
        next: () => {
          this.loading = false;
          const returnUrl = this.route.snapshot.queryParamMap.get('returnUrl') || '/projects';
          this.router.navigateByUrl(returnUrl);
        },
        error: (err: HttpErrorResponse) => {
          this.loading = false;
          this.error = err.error?.error || err.error || 'Неверная почта или пароль';
        }
      });
      return;
    }

    this.authService.register(this.email, this.password).subscribe({
      next: () => {
        this.loading = false;
        this.success = 'Пользователь зарегистрирован. Теперь можно войти.';
        this.mode = 'login';
        this.password = '';
      },
      error: (err: HttpErrorResponse) => {
        this.loading = false;
        this.error = err.error?.error || err.error || 'Не удалось зарегистрировать пользователя';
      }
    });
  }
}

