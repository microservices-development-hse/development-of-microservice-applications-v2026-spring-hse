import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CompPage } from './comp-page';

describe('CompPage', () => {
  let component: CompPage;
  let fixture: ComponentFixture<CompPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [CompPage]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CompPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
