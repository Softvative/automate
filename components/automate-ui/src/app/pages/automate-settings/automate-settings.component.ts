import { Component, OnInit } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { FormGroup, FormBuilder } from '@angular/forms';
import { Store } from '@ngrx/store';
import { NgrxStateAtom } from '../../ngrx.reducers';
import { LayoutFacadeService, Sidebar } from 'app/entities/layout/layout.facade';
import {
  automateSettingsState,
  changeConfiguration
} from '../../entities/automate-settings/automate-settings.selectors';
import {
  GetSettings,
  ConfigureSettings
} from '../../entities/automate-settings/automate-settings.actions';
import {
  JobSchedulerStatus,
  IngestJob,
  IngestJobs
} from '../../entities/automate-settings/automate-settings.model';
import { TelemetryService } from '../../services/telemetry/telemetry.service';

@Component({
  templateUrl: './automate-settings.component.html',
  styleUrls: ['./automate-settings.component.scss']
})

export class AutomateSettingsComponent implements OnInit {
  private defaultFormData = {
    eventFeed: {
      unit: 'd',
      threshold: '3',
      disable: false
    },
    clientRuns: {
      unit: 'd',
      threshold: '',
      disable: false
    },
    complianceData: {
      unit: 'd',
      threshold: '',
      disable: false
    },
    missingNodes: {
      unit: 'd',
      threshold: '',
      disable: false
    },
    deleteMissingNodes: {
      unit: 'd',
      threshold: '',
      disable: false
    }
  };
  eventFeedForm: FormGroup;
  serviceGroupForm: FormGroup;

  clientRunsForm: FormGroup;
  complianceDataForm: FormGroup;
  missingNodesForm: FormGroup;
  deleteMissingNodesForm: FormGroup;
  automateSettingsForm: FormGroup;
  jobSchedulerStatus: JobSchedulerStatus;

  // Has the form changed?
  formChanged = false;

  // Notification bits
  notificationVisible = false;
  notificationType = 'info';
  notificationMessage = 'All settings have been updated successfully';

  constructor(
    private store: Store<NgrxStateAtom>,
    private layoutFacade: LayoutFacadeService,
    private fb: FormBuilder,
    private telemetryService: TelemetryService
  ) {
    const formDetails = this.defaultFormData;
    // this.eventFeedForm = this.fb.group(formDetails['eventFeed']);
    // this.clientRunsForm = this.fb.group(formDetails['clientRuns']);
    // this.complianceDataForm = this.fb.group(formDetails['complianceData']);
    this.missingNodesForm = this.fb.group(formDetails['missingNodes']);
    this.deleteMissingNodesForm = this.fb.group(formDetails['deleteMissingNodes']);

    this.eventFeedForm = this.fb.group({
      feedData: this.fb.group({
        threshold: {value: 3, disabled: true},
        unit: {value: 'd', disabled: true}
      }),
      serverActions: this.fb.group({
        threshold: { value: 30, disabled: true },
        unit: { value: 'd', disabled: true }
      })
    });

    this.serviceGroupForm = this.fb.group({
      healthChecks: this.fb.group({
        threshold: {value: 5, disabled: true},
        unit: { value: 'd', disabled: true }
      }),
      removeServices: this.fb.group({
        threshold: {value: 7, disabled: true},
        unit: { value: 'd', disabled: true }
      })
    });

    this.clientRunsForm = this.fb.group({
      removeData: this.fb.group({
        threshold: {value: 30, disabled: true}
      }),
      labelMissing: this.fb.group({
        threshold: { value: 30, disabled: true },
        unit: { value: 'd', disabled: true }
      }),
      removeMissing: this.fb.group({
        threshold: { value: 30, disabled: true },
        unit: { value: 'd', disabled: true }
      })
    });

    this.complianceDataForm = this.fb.group({
      reports: this.fb.group({
        threshold: { value: 30, disabled: true }
      }),
      scans: this.fb.group({
        threshold: { value: 30, disabled: true }
      })
    });


    this.automateSettingsForm = this.fb.group({
      eventFeed: this.eventFeedForm,
      serviceGroup: this.serviceGroupForm,
      clientRuns: this.clientRunsForm,
      complianceData: this.complianceDataForm

      // now part of client runs -- prob get deleted
      // missingNodes: this.missingNodesForm,
      // deleteMissingNodes: this.deleteMissingNodesForm
    });
  }

  ngOnInit() {
    this.layoutFacade.showSidebar(Sidebar.Settings);
    this.store.dispatch(new GetSettings({}));
    this.store.select(automateSettingsState)
      .subscribe((automateSettingsSelector) => {
        if (automateSettingsSelector.errorResp !== null) {
          const error = automateSettingsSelector.errorResp;
          const errMsg = 'Unable to load settings.';
          this.showErrorNotification(error, errMsg);
        } else {
          this.jobSchedulerStatus = automateSettingsSelector.jobSchedulerStatus;
          this.telemetryService.track('lifecycleConfiguration', this.jobSchedulerStatus);
          this.updateForm(this.jobSchedulerStatus);
          this.onChanges();
        }
      });
  }

  testPrintValues() {
    console.log(this.automateSettingsForm.value);
  }

  // Has the form changed?
  public noChanges() {
    return !this.formChanged;
  }

  public toggleInput(form, _formGroupName: string, checked: boolean) {
    // currentInput is a reference to the input associated with its checkbox
    const currentInput = form.get(_formGroupName);
    checked === true
      ? currentInput.enable()
      : currentInput.disable();
  }

  // Apply the changes that the user updated in the forms
  public applyChanges() {
    // (@afiune) At the moment the only two forms that are enabled are:
    // => 'missingNodes'
    // => 'deleteMissingNodes'
    //
    // We will apply the changes on the rest when we expose the forms
    const jobs: IngestJob[] = [
      IngestJobs.MissingNodes,
      IngestJobs.MissingNodesForDeletion
    ].map(jobName => {
      const jobForm = this.getJobForm(jobName);
      const job = new IngestJob(null);
      job.name = jobName;
      job.running = !jobForm.disable;
      job.threshold = jobForm.threshold + jobForm.unit;
      return job;
    });

    this.store.dispatch(new ConfigureSettings({jobs: jobs}));
    this.store.select(changeConfiguration)
      .subscribe((changeConfigurationSelector) => {
        if (changeConfigurationSelector.errorResp !== null) {
          const error = changeConfigurationSelector.errorResp;
          const errMsg = 'Unable to update one or more settings.';
          this.showErrorNotification(error, errMsg);
        } else {
          this.formChanged = false;
          this.showSuccessNotification();
        }
      });
  }

  // Hides the notification banner
  public hideNotification() {
    this.notificationVisible = false;
  }

  // Returns timeout depending on the type of notification is being displayed
  public notificationTimeout() {
    switch (this.notificationType) {
      case 'error':
        return 8;
      default:
        return 4;
    }
  }

  // Subscribes to any change inside the automateSettingsForm
  private onChanges() {
    this.automateSettingsForm.valueChanges
      .subscribe(_change => this.formChanged = true);
  }

  private showErrorNotification(error: HttpErrorResponse, msg: string) {
    // Extract the error message from the HttpErrorResponse
    // if it is available inside the body.
    //
    // The format looks similar like:
    // {
    //   "error":"invalid time unit: '-1d'",
    //   "code":3,
    //   "details":[]
    // }
    if (error.error.error) {
      msg = msg + ' [' + error.error.error + ']';
    }
    console.error(error.message);
    this.notificationType = 'error';
    this.notificationMessage = msg;
    this.notificationVisible = true;
  }

  private showSuccessNotification() {
    this.notificationType = 'info';
    this.notificationMessage = 'All settings have been updated successfully';
    this.notificationVisible = true;
  }

  // Update forms until we get the job scheduler status
  public updateForm(jobSchedulerStatus: JobSchedulerStatus) {
    if (jobSchedulerStatus === null) {
      return;
    }

    jobSchedulerStatus.jobs.forEach((job: IngestJob) => {
      const [threshold, unit] = this.splitThreshold(job.threshold);
      const form = {
        disable: !job.running,
        threshold: threshold,
        unit: unit
      };

      switch (job.name) {
        case IngestJobs.MissingNodes: {
          this.missingNodesForm = this.fb.group(form);
          break;
        }
        case IngestJobs.MissingNodesForDeletion: {
          this.deleteMissingNodesForm = this.fb.group(form);
          break;
        }
        // TODO @afiune missing forms to add, at the moment we can't modify
        // this parameter/settings since the services take it at startup.
        // (we need to change that first)
        //
        // this.clientRunsForm = this.fb.group(form);
        // this.complianceDataForm = this.fb.group(form);
        // this.eventFeedForm = this.fb.group(form);
      }
    });

    this.automateSettingsForm = this.fb.group({
      eventFeed: this.eventFeedForm,
      serviceGroup: this.serviceGroupForm,
      clientRuns: this.clientRunsForm,
      complianceData: this.complianceDataForm

      // now part of client runs -- prob get deleted
      // missingNodes: this.missingNodesForm,
      // deleteMissingNodes: this.deleteMissingNodesForm
    });
  }

  private getJobForm(job: string) {
    switch (job) {
      case IngestJobs.MissingNodes: {
        return this.automateSettingsForm.value['missingNodes'];
      }
      case IngestJobs.MissingNodesForDeletion: {
        return this.automateSettingsForm.value['deleteMissingNodes'];
      }
    }
  }

  private splitThreshold(threshold: string) {
    return [
      threshold.slice(0, threshold.length - 1),
      threshold.slice(-1)
    ];
  }
}
