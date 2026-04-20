import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

import { IProj } from '../models/proj.model';
import { CheckedProject } from '../models/check-element.model';
import { DatabaseProjectServices } from '../services/database-project.services';
import { ConfigurationService } from '../services/configuration.services';

@Component({
  selector: 'app-compare-page',
  templateUrl: './compare-page.component.html',
  styleUrls: ['./compare-page.component.css']
})
export class ComparePageComponent implements OnInit {
  projects: IProj[] = [];
  checked: Map<string, number> = new Map<string, number>();
  noProjects = false;
  inited = false;

  webUrl = '';

  constructor(
    private configurationService: ConfigurationService,
    private myProjectService: DatabaseProjectServices,
    private router: Router
  ) {
    this.webUrl = this.configurationService.getValue('webUrl') || '';
  }

  ngOnInit(): void {
    this.myProjectService.getAll().subscribe({
      next: projects => {
        this.noProjects = projects.projects.length === 0;
        this.projects = projects.projects;
        this.inited = true;
      },
      error: () => {
        this.noProjects = true;
        this.inited = true;
      }
    });
  }

  childOnChecked(project: CheckedProject): void {
    const key = String(project.Name || '');
    const id = Number(project.Id);

    if (!key) {
      return;
    }

    if (project.Checked) {
      this.checked.set(key, id);
    } else {
      this.checked.delete(key);
    }
  }

  onClickCompare(): void {
    const items: string[] = [];
    const ids: number[] = [];

    this.checked.forEach((value: number, key: string) => {
      if (value !== null && value !== undefined) {
        items.push(key);
        ids.push(value);
      }
    });

    if (items.length > 3) {
      this.showErrorMessage('Максимальное число проектов 3');
      return;
    }

    if (items.length <= 1) {
      this.showErrorMessage('Минимальное число проектов для сравнения 2.');
      return;
    }

    this.router.navigate(['/compare-projects'], {
      queryParams: {
        keys: items,
        value: ids
      }
    });
  }

  showErrorMessage(msg: string): void {
    alert(msg);
  }
}
