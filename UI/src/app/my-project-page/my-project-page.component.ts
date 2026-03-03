import { Component, OnInit } from '@angular/core';
import {DatabaseProjectServices} from "../services/database-project.services";
import {IProj} from "../models/proj.model";
import {CheckedProject} from "../models/check-element.model";
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-my-project-page',
  templateUrl: './my-project-page.component.html',
  styleUrls: ['./my-project-page.component.css']
})
export class MyProjectPageComponent implements OnInit {
  myProjects: IProj[] = []
  checked: Map<any, any> = new Map();
  loading = false
  noProjects: boolean = false
  inited: boolean = false

  constructor(private myProjectService: DatabaseProjectServices) { }

  ngOnInit(): void {
    this.loading = true
    this.myProjectService.getAll().subscribe({
        next: projects => {
            this.noProjects = projects.data.length == 0;
            this.myProjects = projects.data
            this.loading = false
            this.inited = true
        },
        error: (err: HttpErrorResponse) => {
            console.error('MyProjectPage error:', err);
            if (err.status === 0) {
                alert('Сервер недоступен.');
            } else if (err.status === 404) {
                alert('Данные не найдены.');
            } else if (err.status === 500) {
                alert('Ошибка при анализе проекта.');
            } else {
                alert(err.error?.message || 'Неизвестная ошибка');
            }
            this.loading = false;
        }
    })
  }

  childOnChecked(project: CheckedProject){
    if (project.Checked) {
      this.checked.set(project.Name, project.Id)
    }else if (this.checked.has(project.Name)){
      this.checked.delete(project.Name)
    }
  }

}
