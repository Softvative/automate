import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { StoreModule } from '@ngrx/store';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { FormGroup, FormBuilder, FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { ngrxReducers, runtimeChecks } from 'app/ngrx.reducers';
import {
  IngestJob,
  JobSchedulerStatus,
  JobCategories
} from 'app/entities/automate-settings/automate-settings.model';

import { FeatureFlagsService } from 'app/services/feature-flags/feature-flags.service';
import { TelemetryService } from '../../services/telemetry/telemetry.service';
import { AutomateSettingsComponent } from './automate-settings.component';

import { using } from 'app/testing/spec-helpers';

let mockJobSchedulerStatus: JobSchedulerStatus = null;

class MockTelemetryService {
  track() { }
}

// A reusable list of all the form names
const ALL_FORMS = [
  'eventFeedRemoveData',
  'eventFeedServerActions',
  'serviceGroupNoHealthChecks',
  'serviceGroupRemoveServices',
  'clientRunsRemoveData',
  'clientRunsLabelMissing',
  'clientRunsRemoveNodes',
  'complianceRemoveReports',
  'complianceRemoveScans'
];

describe('AutomateSettingsComponent', () => {
  let component: AutomateSettingsComponent;
  let fixture: ComponentFixture<AutomateSettingsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        FormsModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatSelectModule,
        BrowserAnimationsModule,
        HttpClientTestingModule,
        StoreModule.forRoot(ngrxReducers, { runtimeChecks })
      ],
      declarations: [
        AutomateSettingsComponent
      ],
      providers: [
        FormBuilder,
        FeatureFlagsService,
        { provide: TelemetryService, useClass: MockTelemetryService }
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AutomateSettingsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  afterEach(() => {
    fixture.destroy();
  });

  it('exists', () => {
    expect(component).toBeTruthy();
  });

  it('sets defaults for all form groups', () => {
    expect(component.automateSettingsForm).not.toEqual(null);
    expect(component.automateSettingsForm instanceof FormGroup).toBe(true);
    expect(Object.keys(component.automateSettingsForm.controls)).toEqual(ALL_FORMS);
  });

  describe('toggleInput(form, value)', () => {

    using(ALL_FORMS
        // Service Groups on not currently uncheckable through the UI
        .filter( form => !['serviceGroupNoHealthChecks', 'serviceGroupRemoveServices']
        .includes(form)),
        function( form: string) {
      it(`deactivates the associated ${form} form`, () => {
        expect(component[form].value.disabled).toEqual(false);
        component.toggleInput(component[form], false);
        expect(component[form].value.disabled).toEqual(true);
        expect(component[form].get('unit').disabled).toBe(true);
        expect(component[form].get('threshold').disabled).toBe(true);
      });
    });

    using(ALL_FORMS
        // Service Groups on not currently uncheckable through the UI
        .filter( form => !['serviceGroupsNoHealthChecks', 'serviceGroupRemoveServices']
        .includes(form)),
        function (form: string) {
          it(`activates the associated ${form} form`, () => {
        component[form].patchValue({disabled: true}); // Deactivate form to start
        expect(component[form].value.disabled).toEqual(true);
        component.toggleInput(component[form], true);
        expect(component[form].value.disabled).toEqual(false);
        expect(component[form].get('unit').disabled).toBe(false);
        expect(component[form].get('threshold').disabled).toBe(false);
      });
    });

  });

  describe('when jobSchedulerStatus is null', () => {
    it('does not update the forms', () => {
      const formBeforeUpdate = component.automateSettingsForm;
      component.updateForm(null);
      expect(component.automateSettingsForm).toEqual(formBeforeUpdate);
    });
  });

  describe('when jobSchedulerStatus is set', () => {
    beforeAll(() => {
      const eventFeedRemoveData: IngestJob = {
        name: 'periodic_purge',
        category: JobCategories.EventFeed,
        disabled: true,
        threshold: '',
        purge_policies: {
          elasticsearch: [
            {
              name: 'feed',
              older_than_days: 53,
              disabled: false
            }
          ]
        }
      };

      const infraNestedForms: IngestJob = {
        name: 'periodic_purge_timeseries',
        category: JobCategories.Infra,
        disabled: false,
        threshold: '',
        purge_policies: {
          elasticsearch: [
            {
              name: 'actions',
              older_than_days: 22, // default is 30, since disabled
              disabled: true       // is true older than should be null
            },
            {
              name: 'converge-history',
              older_than_days: 12,
              disabled: false
            }
          ]
        }
      };

      const complianceForms: IngestJob = {
        category: JobCategories.Compliance,
        name: 'periodic_purge',
        threshold: '',
        disabled: true,
        purge_policies: {
          elasticsearch: [
            {
              name: 'compliance-reports',
              older_than_days: 105,
              disabled: false
            },
            {
              name: 'compliance-scans',
              older_than_days: 92,
              disabled: false
            }
          ]
        }
      };

      const clientRunsRemoveData: IngestJob = {
        category: JobCategories.Infra,
        name: 'missing_nodes',
        disabled : false,
        threshold : '7d'
      };

      const clientRunsLabelMissing: IngestJob = {
        category: JobCategories.Infra,
        name: 'missing_nodes_for_deletion',
        disabled: false,
        threshold: '14m'
      };

      mockJobSchedulerStatus = new JobSchedulerStatus([
        eventFeedRemoveData,
        infraNestedForms,
        clientRunsRemoveData,
        clientRunsLabelMissing,
        complianceForms
      ]);
    });

    using([
      ['eventFeedRemoveData', false, 53 ],
      ['eventFeedServerActions', true, undefined ], // Infra purge_timeseries -> server_actions
      // ['serviceGroupNoHealthChecks', false, 5 ], // Services not enabled yet
      // ['serviceGroupRemoveServices', false, 5 ], // Services not enabled yet
      ['clientRunsRemoveData', false, '7' ], // Infra Remove data
      ['clientRunsLabelMissing', false, '14' ], // Infra label as missing data
      ['clientRunsRemoveNodes', false, 12 ], // Infra purge_timeseries -> converge_history
      ['complianceRemoveReports', false, 105 ], // Compliance
      ['complianceRemoveScans', false, 92 ] // Compliance
    ], function(formName: string, disabledStatus: boolean, threshold: number | string) {
      it(`when form nested, it updates the ${formName} form group correctly`, () => {
        component.updateForm(mockJobSchedulerStatus);

        const newFormValues = component[formName].value;

        expect(newFormValues.disabled).toEqual(disabledStatus);
        expect(newFormValues.threshold).toEqual(threshold);
      });
    });

    describe('when user applyChanges()', () => {
      it('saves settings', () => {
        component.updateForm(mockJobSchedulerStatus);
        component.applyChanges();

        // expect(component.notificationVisible).toBe(true);
        expect(component.notificationType).toEqual('info');
        expect(component.notificationMessage)
        .toEqual('All settings have been updated successfully');
        expect(component.formChanged).toEqual(false);
        // expect(component.saving).toEqual(false);
      });

      xdescribe('and there is an error', () => {
        it('triggers a notification error (shows a banner)', () => {
          component.updateForm(mockJobSchedulerStatus);
          component.applyChanges();
          expect(component.notificationType).toEqual('error');
          expect(component.notificationMessage)
            .toEqual('Unable to update one or more settings. Verify the console logs.');
          expect(component.notificationVisible).toEqual(true);
        });
      });

    });

  });
});
