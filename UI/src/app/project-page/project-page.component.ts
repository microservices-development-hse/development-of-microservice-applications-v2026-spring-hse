import { Component, OnInit } from '@angular/core';
import {ProjectServices} from "../services/project.services";
import {IProj} from "../models/proj.model";
import {PageInfo} from "../models/pageInfo.model";
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-project-page',
  templateUrl: './project-page.component.html',
  styleUrls: ['./project-page.component.css']
})
export class ProjectPageComponent implements OnInit {
  projects: IProj[] = []
  loading = false
  searchName = ''
  pageInfo: PageInfo
  start_page = 1

  constructor(private projectService: ProjectServices) {
  }

  ngOnInit(): void {
    this.loading = true
    this.projectService.getAll(this.start_page, this.searchName).subscribe({
        next: projects => {
            this.projects = projects.data
            this.loading = false
            this.pageInfo = projects.pageInfo
        },
        error: (err: HttpErrorResponse) => {
            console.error('ProjectPage error:', err);
            if (err.status === 0) {
                alert('Backend недоступен. Проверьте, запущен ли сервер.');
            } else if (err.status === 400) {
                alert('Некорректный запрос.');
            } else if (err.status === 404) {
                alert('Проект не найден.');
            } else if (err.status === 500) {
                alert('Ошибка сервера.');
            } else {
                alert(`Ошибка: ${err.error?.message || err.message}`);
            }
            this.loading = false;
        }
    })
  }

  gty(page: any){
    this.projectService.getAll(page, this.searchName).subscribe({
        next: projects => {
            this.projects = projects.data
            this.pageInfo = projects.pageInfo
            this.loading = false
        },
        error: (err: HttpErrorResponse) => {
            console.error('Error loading page:', err);
            this.loading = false;
        }
    })
  }

  getSearchProjects() {
    this.pageInfo.currentPage = this.start_page;
    this.gty(this.pageInfo.currentPage);
  }
}
