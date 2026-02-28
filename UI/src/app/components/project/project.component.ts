import {Component, Input, OnInit} from '@angular/core'
import {IProj} from "../../models/proj.model";
import {ProjectServices} from "../../services/project.services";

@Component({
  selector: 'app-project',
  templateUrl: './project.component.html',
  styleUrls: ['./project.component.css']
})
export class ProjectComponent implements OnInit {
  @Input() project: IProj
  adding: Boolean;

  constructor(private projectService: ProjectServices) {
    //TO_DO
  }

  ngOnInit(): void {
    this.adding = this.project.Existence;
  }

  addMyProject(project: IProj) {
    if (!this.adding) {
        this.projectService.addProject(project.Key).subscribe({
            next: resp => {
                this.adding = !this.adding
            },
            error: error => {
                if (error.status == 0){
                    alert("Unable to connect to backend")
                }
                if (error.status == 400){
                    alert(error.error?.message || error.message)
                }
            }
        });
    } else {
        console.log(this.project.Id);
        this.projectService.deleteProject(project.Id).subscribe({
            next: resp => {
                this.adding = !this.adding
            },
            error: error => {
                if (error.status == 0){
                    alert("Unable to connect to backend")
                }
                if (error.status == 400){
                    alert("Unable to connect to DB")
                }
            }
        });
    }
  }
}

