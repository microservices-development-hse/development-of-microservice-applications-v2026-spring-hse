import { Component, Input, OnInit } from '@angular/core';
import { IProj } from '../../models/proj.model';
import { ProjectServices } from '../../services/project.services';

@Component({
  selector: 'app-project',
  templateUrl: './project.component.html',
  styleUrls: ['./project.component.css']
})
export class ProjectComponent implements OnInit {
  @Input() project!: IProj;
  adding = false;

  constructor(private projectService: ProjectServices) {}

  ngOnInit(): void {
    this.adding = false;
  }

  addMyProject(project: IProj) {
    if (!this.adding) {
      this.projectService.addProject(project).subscribe({
        next: () => {
          this.adding = !this.adding;
        },
        error: error => {
          if (error.status == 0) {
            alert("Unable to connect to backend");
          }
          if (error.status == 400) {
            alert(error.error?.message || error.message);
          }
        }
      });
    } else {
      this.projectService.deleteProject(project.Id).subscribe({
        next: () => {
          this.adding = !this.adding;
        },
        error: error => {
          if (error.status == 0) {
            alert("Unable to connect to backend");
          }
          if (error.status == 400) {
            alert("Unable to connect to DB");
          }
        }
      });
    }
  }
}
