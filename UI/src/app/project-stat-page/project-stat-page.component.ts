import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from "@angular/router";
import { DatabaseProjectServices } from "../services/database-project.services";
import { Chart } from "angular-highcharts";
import { openTaskChartOptions } from "./helpers/openTaskChartOptions";
import { openStateChartOptions } from "./helpers/openStateChartOptions";
import { resolveStateChartOptions } from "./helpers/resolveStateChartOptions";
import { progressStateChartOptions } from "./helpers/progressStateChartOptions";
import { reopenStateChartOptions } from "./helpers/reopenStateChartOptions";
import { activityByTaskChartOptions } from "./helpers/activityByTaskChartOptions";
import { taskPriorityChartOptions } from "./helpers/taskPriorityChartOptions";
import { closeTaskPriorityChartOptions } from "./helpers/closeTaskPriorityChartOptions";
import { complexityTaskChartOptions } from "./helpers/complexityTaskChartOptions";
import { ConfigurationService } from "../services/configuration.services";

@Component({
  selector: 'project-stat-page',
  templateUrl: './project-stat-page.component.html',
  styleUrls: ['./project-stat-page.component.css']
})
export class ProjectStatPageComponent implements OnInit {
  projects: string[] = [];
  selectedCharts: number[] = [];
  projectId = 0;

  openTaskChart = new Chart();
  openStateChart = new Chart();
  resolveStateChart = new Chart();
  progressStateChart = new Chart();
  reopenStateChart = new Chart();
  complexityTaskChart = new Chart();
  activityByTaskChart = new Chart();
  taskPriorityChart = new Chart();
  closeTaskPriorityChart = new Chart();

  webUrl = "";

  constructor(
    private configurationService: ConfigurationService,
    private route: ActivatedRoute,
    private dbProjectService: DatabaseProjectServices
  ) {
    this.projects = this.route.snapshot.queryParamMap.getAll("keys");
    this.selectedCharts = this.route.snapshot.queryParamMap
      .getAll("value")
      .map(v => Number(v))
      .filter(v => !isNaN(v));

    const projectIdParam = this.route.snapshot.queryParamMap.get("projectId");
    this.projectId = projectIdParam ? Number(projectIdParam) : 0;

    this.webUrl = configurationService.getValue("webUrl");
  }

  ngOnInit(): void {
    if (!this.projectId) {
      return;
    }

    const projectId = this.projectId.toString();
    const selected = new Set(this.selectedCharts);

    if (selected.has(1)) {
      this.dbProjectService.getAnalytics(projectId, "bottlenecks").subscribe({
        next: info => this.renderOpenTaskChart(info),
        error: () => this.removeChart('open-task', 'open-task-title')
      });
    } else {
      this.removeChart('open-task', 'open-task-title');
    }

    if (selected.has(2)) {
      this.dbProjectService.getAnalytics(projectId, "life_cycle").subscribe({
        next: info => this.renderLifecycleCharts(info),
        error: () => this.removeLifecycleCharts()
      });
    } else {
      this.removeLifecycleCharts();
    }

    if (selected.has(3) || selected.has(4)) {
      this.dbProjectService.getAnalytics(projectId, "complexity").subscribe({
        next: info => {
          if (selected.has(3)) {
            this.renderActivityByTaskChart(info);
          } else {
            this.removeChart('activity-by-task', 'activity-by-task-title');
          }

          if (selected.has(4)) {
            this.renderComplexityTaskChart(info);
          } else {
            this.removeChart('complexity-task', 'complexity-task-title');
          }
        },
        error: () => {
          this.removeChart('activity-by-task', 'activity-by-task-title');
          this.removeChart('complexity-task', 'complexity-task-title');
        }
      });
    } else {
      this.removeChart('activity-by-task', 'activity-by-task-title');
      this.removeChart('complexity-task', 'complexity-task-title');
    }

    if (selected.has(5) || selected.has(6)) {
      this.dbProjectService.getAnalytics(projectId, "priority").subscribe({
        next: info => {
          if (selected.has(5)) {
            this.renderPriorityChart(info, taskPriorityChartOptions, 'taskPriorityChart', 'task-priority', 'task-priority-title');
          } else {
            this.removeChart('task-priority', 'task-priority-title');
          }

          if (selected.has(6)) {
            this.renderPriorityChart(info, closeTaskPriorityChartOptions, 'closeTaskPriorityChart', 'close-task-priority', 'close-task-priority-title');
          } else {
            this.removeChart('close-task-priority', 'close-task-priority-title');
          }
        },
        error: () => {
          this.removeChart('task-priority', 'task-priority-title');
          this.removeChart('close-task-priority', 'close-task-priority-title');
        }
      });
    } else {
      this.removeChart('task-priority', 'task-priority-title');
      this.removeChart('close-task-priority', 'close-task-priority-title');
    }
  }

  private projectName(): string {
    return this.projects[0] ? this.projects[0].toString() : 'Project';
  }

  private removeChart(containerId: string, titleId: string): void {
    const chartElem = document.getElementById(containerId);
    const titleElem = document.getElementById(titleId);

    if (chartElem) {
      chartElem.remove();
    }
    if (titleElem) {
      titleElem.remove();
    }
  }

  private removeLifecycleCharts(): void {
    this.removeChart('open-state', 'open-state-title');
    this.removeChart('resolve-state', 'resolve-state-title');
    this.removeChart('progress-state', 'progress-state-title');
    this.removeChart('reopen-state', 'reopen-state-title');
  }

  private buildHistogram(values: number[], bins = 8): { categories: string[]; counts: number[] } {
    const clean = values.filter(v => !isNaN(v) && isFinite(v));

    if (clean.length === 0) {
      return { categories: [], counts: [] };
    }

    const min = Math.min(...clean);
    const max = Math.max(...clean);

    if (min === max) {
      return {
        categories: [min.toFixed(1)],
        counts: [clean.length]
      };
    }

    const step = (max - min) / bins;
    const counts = new Array(bins).fill(0);

    clean.forEach(value => {
      let idx = Math.floor((value - min) / step);
      if (idx >= bins) {
        idx = bins - 1;
      }
      if (idx < 0) {
        idx = 0;
      }
      counts[idx]++;
    });

    const categories = counts.map((_, i) => {
      const from = min + step * i;
      const to = min + step * (i + 1);
      return `${from.toFixed(1)}-${to.toFixed(1)}`;
    });

    return { categories, counts };
  }

  private getCaseInsensitiveValue(source: any, names: string[]): any {
    if (!source) {
      return undefined;
    }

    const keys = Object.keys(source);
    for (const name of names) {
      const found = keys.find(k => k.toLowerCase() === name.toLowerCase());
      if (found) {
        return source[found];
      }
    }

    return undefined;
  }

  private renderOpenTaskChart(info: any): void {
    const data = Array.isArray(info) ? info : [];

    if (data.length === 0) {
      this.removeChart('open-task', 'open-task-title');
      return;
    }

    const categories = data.map((item: any) => item.issue_key ?? item.issueKey ?? item.key ?? '');
    const counts = data.map((item: any) => Number(item.time_in_status ?? item.timeInStatus ?? 0));

    openTaskChartOptions.xAxis = {
      ...(openTaskChartOptions.xAxis as object),
      categories
    } as any;

    openTaskChartOptions.series = [
      {
        type: 'column',
        name: this.projectName(),
        data: counts
      } as any
    ];

    this.openTaskChart = new Chart(openTaskChartOptions);
  }

  private renderLifecycleCharts(info: any): void {
    const openValues = this.getCaseInsensitiveValue(info, ['open']);
    const resolveValues = this.getCaseInsensitiveValue(info, ['resolve', 'resolved']);
    const progressValues = this.getCaseInsensitiveValue(info, ['progress', 'in progress', 'in_progress']);
    const reopenValues = this.getCaseInsensitiveValue(info, ['reopen']);

    this.renderLifecycleChart(openValues, openStateChartOptions, 'openStateChart', 'open-state', 'open-state-title');
    this.renderLifecycleChart(resolveValues, resolveStateChartOptions, 'resolveStateChart', 'resolve-state', 'resolve-state-title');
    this.renderLifecycleChart(progressValues, progressStateChartOptions, 'progressStateChart', 'progress-state', 'progress-state-title');
    this.renderLifecycleChart(reopenValues, reopenStateChartOptions, 'reopenStateChart', 'reopen-state', 'reopen-state-title');
  }

  private renderLifecycleChart(values: any, chartOptions: any, field: string, containerId: string, titleId: string): void {
    if (!Array.isArray(values) || values.length === 0) {
      this.removeChart(containerId, titleId);
      return;
    }

    const numbers = values.map((v: any) => Number(v)).filter((v: number) => !isNaN(v) && isFinite(v));
    const histogram = this.buildHistogram(numbers, 8);

    if (histogram.categories.length === 0) {
      this.removeChart(containerId, titleId);
      return;
    }

    chartOptions.xAxis = {
      ...(chartOptions.xAxis as object),
      categories: histogram.categories
    } as any;

    chartOptions.series = [
      {
        type: 'column',
        name: this.projectName(),
        data: histogram.counts
      } as any
    ];

    (this as any)[field] = new Chart(chartOptions);
  }

  private renderActivityByTaskChart(info: any): void {
    const data = Array.isArray(info) ? info : [];

    if (data.length === 0) {
      this.removeChart('activity-by-task', 'activity-by-task-title');
      return;
    }

    const categories = data.map((item: any) => item.issue_key ?? item.issueKey ?? item.key ?? '');
    const counts = data.map((item: any) => Number(item.move_count ?? item.moveCount ?? 0));

    activityByTaskChartOptions.xAxis = {
      ...(activityByTaskChartOptions.xAxis as object),
      categories
    } as any;

    activityByTaskChartOptions.series = [
      {
        type: 'spline',
        name: this.projectName() + ' moves',
        data: counts
      } as any
    ];

    this.activityByTaskChart = new Chart(activityByTaskChartOptions);
  }

  private renderComplexityTaskChart(info: any): void {
    const data = Array.isArray(info) ? info : [];

    if (data.length === 0) {
      this.removeChart('complexity-task', 'complexity-task-title');
      return;
    }

    const categories = data.map((item: any) => item.issue_key ?? item.issueKey ?? item.key ?? '');
    const values = data.map((item: any) => Number(item.lead_time ?? item.leadTime ?? 0));

    complexityTaskChartOptions.xAxis = {
      ...(complexityTaskChartOptions.xAxis as object),
      categories
    } as any;

    complexityTaskChartOptions.series = [
      {
        type: 'column',
        name: this.projectName(),
        data: values
      } as any
    ];

    this.complexityTaskChart = new Chart(complexityTaskChartOptions);
  }

  private renderPriorityChart(info: any, chartOptions: any, field: string, containerId: string, titleId: string): void {
    const data = Array.isArray(info) ? info : [];

    if (data.length === 0) {
      this.removeChart(containerId, titleId);
      return;
    }

    const categories = data.map((item: any) => item.name ?? item.priority ?? item.key ?? '');
    const values = data.map((item: any) => Number(item.value ?? 0));

    chartOptions.xAxis = {
      ...(chartOptions.xAxis as object),
      categories
    } as any;

    chartOptions.series = [
      {
        type: 'column',
        name: this.projectName(),
        data: values
      } as any
    ];

    (this as any)[field] = new Chart(chartOptions);
  }

  ngOnDestroy(): void {
    // @ts-ignore
    openTaskChartOptions.xAxis["categories"] = [];
    openTaskChartOptions.series = [];

    // @ts-ignore
    openStateChartOptions.xAxis["categories"] = [];
    openStateChartOptions.series = [];

    // @ts-ignore
    resolveStateChartOptions.xAxis["categories"] = [];
    resolveStateChartOptions.series = [];

    // @ts-ignore
    progressStateChartOptions.xAxis["categories"] = [];
    progressStateChartOptions.series = [];

    // @ts-ignore
    reopenStateChartOptions.xAxis["categories"] = [];
    reopenStateChartOptions.series = [];

    // @ts-ignore
    activityByTaskChartOptions.xAxis["categories"] = [];
    activityByTaskChartOptions.series = [];

    // @ts-ignore
    taskPriorityChartOptions.xAxis["categories"] = [];
    taskPriorityChartOptions.series = [];

    // @ts-ignore
    closeTaskPriorityChartOptions.xAxis["categories"] = [];
    closeTaskPriorityChartOptions.series = [];

    // @ts-ignore
    complexityTaskChartOptions.xAxis["categories"] = [];
    complexityTaskChartOptions.series = [];
  }
}
