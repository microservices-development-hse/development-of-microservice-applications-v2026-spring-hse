import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { DatabaseProjectServices } from '../services/database-project.services';
import { Chart } from 'angular-highcharts';
import { openTaskChartOptions } from './helpers/openTaskChartOptions';
import { ConfigurationService } from '../services/configuration.services';

@Component({
  selector: 'app-compare-project-page',
  templateUrl: './compare-project-page.component.html',
  styleUrls: ['./compare-project-page.component.scss']
})
export class CompareProjectPageComponent implements OnInit {
  projects: string[] = [];
  ids: string[] = [];
  resultReq: ReqData[] = [];
  openTaskChart = new Chart();
  webUrl = '';

  constructor(
    private configurationService: ConfigurationService,
    private route: ActivatedRoute,
    private dbProjectService: DatabaseProjectServices
  ) {
    this.projects = this.route.snapshot.queryParamMap.getAll('keys');
    this.ids = this.route.snapshot.queryParamMap.getAll('projectIds');
    this.ids = this.ids
      .map(id => Number(id))
      .filter(id => Number.isInteger(id) && id > 0)
      .map(id => String(id));
    this.webUrl = configurationService.getValue('webUrl');
  }

  ngOnInit(): void {
    if (!this.ids.length) {
      return;
    }

    this.ids = this.ids.filter(id => Number(id) > 0);
    if (!this.ids.length) {
      return;
    }

    forkJoin(
      this.ids.map(id =>
        this.dbProjectService.getProjectStatByID(id).pipe(
          catchError(err => {
            console.error(`Failed to load stat for project ${id}:`, err);
            return of(null);
          })
        )
      )
    ).subscribe(stats => {
      this.resultReq = stats.map((item: any) => item?.data ?? this.emptyStat());
    });

    forkJoin(
      this.ids.map(id =>
        this.dbProjectService.getAnalytics(id, 'bottlenecks').pipe(
          catchError(err => {
            console.error(`Failed to load bottlenecks for project ${id}:`, err);
            return of([]);
          })
        )
      )
    ).subscribe(projectResponses => {
      const projectSeries = projectResponses.map((response: any, index: number) => ({
        name: this.projects[index] ?? `Project ${index + 1}`,
        values: this.extractOpenTaskValues(response)
      }));

      this.renderOpenTaskCompareChart(projectSeries);
    });
  }

  private emptyStat(): ReqData {
    return {
      Id: 0,
      Key: '',
      Name: '',
      allIssuesCount: 0,
      averageIssuesCount: 0 as any,
      averageTime: 0,
      closeIssuesCount: 0,
      openIssuesCount: 0,
      resolvedIssuesCount: 0,
      reopenedIssuesCount: 0,
      progressIssuesCount: 0
    };
  }

  private extractOpenTaskValues(response: any): number[] {
    const data = Array.isArray(response) ? response : [];
    return data
      .map((item: any) => Number(item.time_in_status ?? item.timeInStatus ?? 0))
      .filter((v: number) => !isNaN(v) && isFinite(v) && v >= 0);
  }

  private renderOpenTaskCompareChart(projectSeries: { name: string; values: number[] }[]): void {
    const allValues = projectSeries.flatMap(p => p.values);

    if (allValues.length === 0) {
      const openTaskElem = document.getElementById('open-task');
      const openTaskTitle = document.getElementById('open-task-title');
      openTaskElem?.remove();
      openTaskTitle?.remove();
      return;
    }

    const colors = ['blue', 'green', 'red', 'orange', 'purple', 'black'];

    const min = Math.min(...allValues);
    const max = Math.max(...allValues);
    const bins = Math.min(8, Math.max(1, Math.ceil(Math.sqrt(allValues.length))));
    const step = max === min ? 1 : (max - min) / bins;

    const categories = Array.from({ length: bins }, (_, i) => {
      const from = min + step * i;
      const to = min + step * (i + 1);
      return `${from.toFixed(1)}-${to.toFixed(1)}`;
    });

    const series = projectSeries.map((project, idx) => {
      const counts = new Array<number>(bins).fill(0);

      project.values.forEach(value => {
        const bin = max === min
          ? 0
          : Math.min(bins - 1, Math.floor((value - min) / step));

        counts[bin]++;
      });

      return {
        name: project.name,
        type: 'column',
        color: colors[idx % colors.length],
        data: counts
      } as any;
    });

    // @ts-ignore
    openTaskChartOptions.xAxis['categories'] = categories;
    openTaskChartOptions.series = series;

    this.openTaskChart = new Chart(openTaskChartOptions);
  }

  ngOnDestroy(): void {
    // @ts-ignore
    openTaskChartOptions.xAxis['categories'] = [];
    openTaskChartOptions.series = [];
  }
}

class ReqData {
  Id: number;
  Key: string;
  Name: string;
  allIssuesCount: number;
  averageIssuesCount: string | number;
  averageTime: number;
  closeIssuesCount: number;
  openIssuesCount: number;
  resolvedIssuesCount: number;
  reopenedIssuesCount: number;
  progressIssuesCount: number;
}
