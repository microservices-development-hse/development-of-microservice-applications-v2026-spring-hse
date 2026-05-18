import { CheckedSetting } from '../../models/check-setting.model';
import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { IProj } from "../../models/proj.model";
import { CheckedProject } from "../../models/check-element.model";
import { SettingBox } from "../../models/setting.model";
import { Router } from "@angular/router";
import { DatabaseProjectServices } from "../../services/database-project.services";

@Component({
  selector: 'app-my-project',
  templateUrl: './my-project.component.html',
  styleUrls: ['./my-project.component.css']
})
export class MyProjectComponent implements OnInit {
  @Output() onChecked: EventEmitter<any> = new EventEmitter<{}>();
  @Input() myProject!: IProj;

  stat: ProjectStat = new ProjectStat();
  checked = 0;
  complited = 0;
  processed = false;
  settings = false;
  checkboxes: SettingBox[] = [];
  setting: Map<any, any> = new Map();

  constructor(private router: Router, private dbProjectService: DatabaseProjectServices) {}

  ngOnInit(): void {
    this.processed = false;
    this.settings = false;

    this.checkboxes.push(new SettingBox("Гистограмма, отражающая время, которое задачи провели в открытом состоянии", false, 1));
    this.checkboxes.push(new SettingBox("Диаграммы, которые показывают распределение времени по состоянием задач", false, 2));
    this.checkboxes.push(new SettingBox("График активности по задачам", false, 3));
    this.checkboxes.push(new SettingBox("График сложности задач", false, 4));
    this.checkboxes.push(new SettingBox("График, отражающий приоритетность всех задач", false, 5));
    this.checkboxes.push(new SettingBox("График, отражающий приоритетность закрытых задач", false, 6));

    this.dbProjectService.getProjectStatByID(this.myProject.Id.toString()).subscribe(projects => {
      this.stat.AverageIssuesCount = projects.data["allIssuesCount"];
      this.stat.OpenIssuesCount = projects.data["openIssuesCount"];
      this.stat.AllIssuesCount = projects.data["allIssuesCount"];
      this.stat.AverageTime = projects.data["averageTime"];
      this.stat.CloseIssuesCount = projects.data["closeIssuesCount"];
      this.stat.ReopenedIssuesCount = projects.data["reopenedIssuesCount"];
      this.stat.ResolvedIssuesCount = projects.data["resolvedIssuesCount"];
      this.stat.ProgressIssuesCount = projects.data["progressIssuesCount"];
    });
  }

  processProject(): void {
    this.checked = 0;
    this.complited = 0;

    this.checkboxes.forEach((box: SettingBox) => {
      if (box.Checked) {
        this.checked++;
      }
    });

    if (this.checked === 0) {
      return;
    }

    this.dbProjectService.recalculateProject(this.myProject.Id.toString()).subscribe({
      next: () => {
        this.complited = this.checked;
        this.processed = true;
        this.checkResult();
      },
      error: error => {
        if (error.status === 0) {
          alert("Unable to connect to backend");
        } else {
          alert(error.error?.message || error.message || "Failed to process project");
        }
      }
    });
  }

  checkResult(): void {
    const ids: number[] = [];
    const items = this.myProject.Name;

    this.checkboxes.forEach((box: SettingBox) => {
      if (box.Checked) {
        ids.push(Number(box.BoxId));
      }
    });

    this.router.navigate([`/project-stat`], {
      queryParams: {
        projectId: this.myProject.Id,
        keys: items,
        value: ids
      }
    });
  }

  clickOnSettings(): void {
    this.settings = !this.settings;
  }

  disableCheckResultButton(): boolean {
    return !this.processed || this.checked !== this.complited;
  }

  disableAnalyzeButton(): boolean {
    return !this.checkboxes.some(checkbox => checkbox.Checked) || this.checked !== this.complited;
  }

  childOnChecked(setting: CheckedSetting): void {
    if (setting.Checked) {
      this.setting.set(setting.ProjectName, setting.BoxId);
    } else if (this.setting.has(setting.ProjectName)) {
      this.setting.delete(setting.ProjectName);
    }

    this.checkboxes[Number(setting.BoxId) - 1].Checked = setting.Checked;
  }
}

class ProjectStat {
  AllIssuesCount!: number;
  AverageIssuesCount!: number;
  AverageTime!: number;
  CloseIssuesCount!: number;
  OpenIssuesCount!: number;
  ResolvedIssuesCount!: number;
  ReopenedIssuesCount!: number;
  ProgressIssuesCount!: number;
}
