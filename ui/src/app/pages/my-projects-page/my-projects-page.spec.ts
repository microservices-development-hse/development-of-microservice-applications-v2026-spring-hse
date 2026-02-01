import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MyProjectsPage } from './my-projects-page';

describe('MyProjectsPage', () => {
  let component: MyProjectsPage;
  let fixture: ComponentFixture<MyProjectsPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MyProjectsPage]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MyProjectsPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
