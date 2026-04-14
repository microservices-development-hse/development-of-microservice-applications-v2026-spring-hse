import { Component, OnInit } from '@angular/core';
import {IProj} from "../models/proj.model";
import {DatabaseProjectServices} from "../services/database-project.services"
import { Router } from '@angular/router';
import {CheckedProject} from "../models/check-element.model";
import {ConfigurationService} from "../services/configuration.services";


@Component({
  selector: 'app-compare-page',
  templateUrl: './compare-page.component.html',
  styleUrls: ['./compare-page.component.css']
})
export class ComparePageComponent implements OnInit {
  projects: IProj[] = []
  checked: Map<any, any> = new Map();
  noProjects: boolean = false
  inited: boolean = false

  webUrl = ""
  constructor(private configurationService: ConfigurationService, private myProjectService: DatabaseProjectServices, private router: Router) {
    this.webUrl = configurationService.getValue("webUrl")
  }
  
  ngOnInit(): void {
    this.myProjectService.getAll().subscribe(projects => {
      this.noProjects = projects.projects.length == 0;
      this.projects = projects.projects.map(p => ({
        ...p,
        Existence: false
      }));
      this.inited = true;
    });
  }

  childOnChecked(project: CheckedProject) {
    const key =
      (project as any).Key ??
      (project as any).key ??
      (project as any).project ??
      (project as any).ProjectKey;

    if (!key) {
      return;
    }

    if (project.Checked) {
      this.checked.set(key, project.Checked);
    } else if (this.checked.has(key)) {
      this.checked.delete(key);
    }
  }

  onClickCompare(): void {
    const items: string[] = [];
    const ids: number[] = [];
  
    this.projects.forEach(p => {
      if (this.checked.get(p.Key)) {
        items.push(p.Key);
        ids.push(p.Id);
      }
    });

    if (items.length > 3) {
      this.showErrorMessage("Максимальное число проектов 3");
    } else if (items.length <= 1) {
      this.showErrorMessage("Минимальное число проектов для сравнения 2.");
    } else {
      this.router.navigate([`/compare-projects`], {
        queryParams: {
          keys: items,
          value: ids
        }
      });
    }
  }

  showErrorMessage(msg: string){
    alert(msg)
  }

}
