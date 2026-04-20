import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';

import { IProj } from '../../models/proj.model';
import { CheckedProject } from '../../models/check-element.model';

@Component({
  selector: 'project-checkbox',
  templateUrl: './checkbox-with-project.component.html',
  styleUrls: ['./checkbox-with-project.component.css']
})
export class ProjectWithCheckboxComponent implements OnInit {
  @Output() onChecked: EventEmitter<CheckedProject> = new EventEmitter<CheckedProject>();
  @Input() project!: IProj;

  isChecked = false;

  ngOnInit(): void {
    this.isChecked = this.project.Existence;
  }

  changed(): void {
    this.onChecked.emit(
      new CheckedProject(this.project.Name, this.isChecked, this.project.Id)
    );
  }
}
