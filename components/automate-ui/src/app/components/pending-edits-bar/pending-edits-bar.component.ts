import { Component, OnDestroy, ViewEncapsulation } from '@angular/core';
import { interval as observableInterval, Subject } from 'rxjs';
import { take, takeUntil } from 'rxjs/operators';
import { Store } from '@ngrx/store';
import { NgrxStateAtom } from 'app/ngrx.reducers';

import { LayoutFacadeService } from 'app/entities/layout/layout.facade';
import { ProjectService } from 'app/entities/projects/project.service';
import { Project } from 'app/entities/projects/project.model';
import { allProjects } from 'app/entities/projects/project.selectors';
import { ApplyRulesStatus } from 'app/entities/projects/project.reducer';

@Component({
  selector: 'app-pending-edits-bar',
  templateUrl: './pending-edits-bar.component.html',
  styleUrls: ['./pending-edits-bar.component.scss'],
  encapsulation: ViewEncapsulation.None
})
export class PendingEditsBarComponent implements OnDestroy {
  public confirmApplyStartModalVisible = false;
  private isDestroyed = new Subject<boolean>();
  private projectsHaveStagedChanges = false;
  public updateProjectsFailed = false;
  private updateProjectsCancelled = false;

  constructor(
    public layoutFacade: LayoutFacadeService,
    private store: Store<NgrxStateAtom>,
    public projects: ProjectService
    ) {
      this.projects.getApplyRulesStatus();
      this.store.select(allProjects).pipe(
        takeUntil(this.isDestroyed)
      ).subscribe((projectList: Project[]) => {
        this.projectsHaveStagedChanges = projectList.some(p => p.status === 'EDITS_PENDING');
        this.updateDisplay();
      });

      this.projects.applyRulesStatus$
      .pipe(takeUntil(this.isDestroyed))
      .subscribe(({ failed, cancelled }: ApplyRulesStatus) => {
        this.updateProjectsFailed = failed;
        this.updateProjectsCancelled = cancelled;
        this.updateDisplay();
      });
    }

  ngOnDestroy(): void {
    this.isDestroyed.next(true);
    this.isDestroyed.complete();
  }

  public updateDisplay() {
    this.layoutFacade.layout.userNotifications.pendingEdits =
      this.projectsHaveStagedChanges
      || this.updateProjectsCancelled
      || this.updateProjectsFailed;
    this.layoutFacade.updateDisplay();
  }

  public openConfirmUpdateStartModal(): void {
    this.confirmApplyStartModalVisible = true;
  }

  private closeConfirmApplyStartModal(): void {
    this.confirmApplyStartModalVisible = false;
  }

  public confirmApplyStart(): void {
    this.closeConfirmApplyStartModal();
    this.projects.applyRulesStart();
    // Rapid sampling for 3 seconds for more responsive UX.
    // If the update is still running, the secondary (active) emitter
    // will check this status at frequent intervals.
    // Once the update completes, the tertiary (dormant) emitter
    // will check this status at INfrequent intervals.
    // (See getActiveApplyRulesStatus$ and getDormantApplyRulesStatus$.)
    observableInterval(250).pipe(take(12)) // 12 x 250ms => 3 seconds
      .subscribe(() => {
        this.projects.getApplyRulesStatus();
      });
  }

  public cancelApplyStart(): void {
    this.closeConfirmApplyStartModal();
  }

  get isBarHidden() {
    return !this.layoutFacade.layout.userNotifications.pendingEdits
      || this.layoutFacade.layout.userNotifications.updatesProcessing;
  }
}
