import { APP_INITIALIZER, NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule, Routes } from "@angular/router";
import { HttpClientModule, HTTP_INTERCEPTORS } from "@angular/common/http";
import { FormsModule } from "@angular/forms";
import { NgxPaginationModule } from "ngx-pagination";
import { ChartModule } from "angular-highcharts";

import { AppComponent } from './app.component';
import { HomePageComponent } from './home-page/home-page.component';
import { ProjectComponent } from "./components/project/project.component";
import { ProjectPageComponent } from './project-page/project-page.component';
import { MyProjectPageComponent } from './my-project-page/my-project-page.component';
import { MyProjectComponent } from './components/my-project/my-project.component';
import { ComparePageComponent } from './compare-page/compare-page.component';
import { ProjectWithCheckboxComponent } from "./components/checkbox-with-project/checkbox-with-project.component";
import { CompareProjectPageComponent } from './compare-project-page/compare-project-page.component';
import { CheckboxWithSettingsComponent } from "./components/checkbox-with-settings/checkbox-with-settings.component";
import { ProjectStatPageComponent } from './project-stat-page/project-stat-page.component';
import { LoginPageComponent } from './login-page/login-page.component';
import { LogoutPageComponent } from './logout-page/logout-page.component';

import { ConfigurationService } from "./services/configuration.services";
import { AuthInterceptor } from "./services/auth.interceptor";
import { AuthGuard } from "./services/auth.guard";

export function initApp(configurationService: ConfigurationService) {
  return () => configurationService.load().toPromise();
}

const routes: Routes = [
  { path: '', pathMatch: 'full' as const, redirectTo: 'login' },
  { path: 'login', component: LoginPageComponent },
  { path: 'logout', component: LogoutPageComponent },
  { path: 'projects', component: ProjectPageComponent, canActivate: [AuthGuard] },
  { path: 'compare', component: ComparePageComponent, canActivate: [AuthGuard] },
  { path: 'myprojects', component: MyProjectPageComponent, canActivate: [AuthGuard] },
  { path: 'compare-projects', component: CompareProjectPageComponent, canActivate: [AuthGuard] },
  { path: 'projects-settings', component: CheckboxWithSettingsComponent, canActivate: [AuthGuard] },
  { path: 'project-stat', component: ProjectStatPageComponent, canActivate: [AuthGuard] },
  { path: '**', redirectTo: 'login' }
];

@NgModule({
  declarations: [
    AppComponent,
    HomePageComponent,
    ProjectComponent,
    ProjectPageComponent,
    MyProjectComponent,
    MyProjectPageComponent,
    ComparePageComponent,
    ProjectWithCheckboxComponent,
    CompareProjectPageComponent,
    CheckboxWithSettingsComponent,
    ProjectStatPageComponent,
    LoginPageComponent,
    LogoutPageComponent
  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    RouterModule.forRoot(routes),
    FormsModule,
    NgxPaginationModule,
    ChartModule
  ],
  providers: [
    {
      provide: APP_INITIALIZER,
      useFactory: initApp,
      multi: true,
      deps: [ConfigurationService]
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }

