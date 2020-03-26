import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { StoreModule } from '@ngrx/store';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { FormGroup, FormBuilder, FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule, MatSelectModule, MatOptionModule } from '@angular/material';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { ngrxReducers, runtimeChecks } from 'app/ngrx.reducers';
import {
  IngestJob,
  IngestJobs,
  JobSchedulerStatus
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
        MatOptionModule,
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

    using([
      ['eventFeedRemoveData', false, true],
      ['eventFeedRemoveData', false, true ],
      ['eventFeedServerActions', false, true ],
      // Service Groups on not currently uncheckable through the UI
      // ['serviceGroupNoHealthChecks', false, true ],
      // ['serviceGroupRemoveServices', false, true ],
      ['clientRunsRemoveData', false, true ],
      ['clientRunsLabelMissing', false, true ],
      ['clientRunsRemoveNodes', false, true ],
      ['complianceRemoveReports', false, true ],
      ['complianceRemoveScans', false, true ]
    ], function( form: string, currentState: boolean, expectedState: boolean) {
      it('deactivates the asociated form', () => {
        expect(component[form].value.disabled).toEqual(currentState);
        component.toggleInput(component[form], currentState);
        expect(component[form].value.disabled).toEqual(expectedState);
        expect(component[form].get('unit').disabled).toBe(expectedState);
        expect(component[form].get('threshold').disabled).toBe(expectedState);
      });
    });

    using([
      ['eventFeedRemoveData', true, false ],
      ['eventFeedRemoveData', true, false ],
      ['eventFeedServerActions', true, false ],
      // Service Groups on not currently uncheckable through the UI
      // ['serviceGroupNoHealthChecks', true, false ],
      // ['serviceGroupRemoveServices', true, false ],
      ['clientRunsRemoveData', true, false ],
      ['clientRunsLabelMissing', true, false ],
      ['clientRunsRemoveNodes', true, false ],
      ['complianceRemoveReports', true, false ],
      ['complianceRemoveScans', true, false ]
    ], function (form: string, currentState: boolean, expectedState: boolean) {
      it('deactivates the asociated form', () => {
        component[form].patchValue({disabled: true}); // Set each form to Activated to start
        expect(component[form].value.disabled).toEqual(currentState);
        component.toggleInput(component[form], currentState);
        expect(component[form].value.disabled).toEqual(expectedState);
        expect(component[form].get('unit').disabled).toBe(expectedState);
        expect(component[form].get('threshold').disabled).toBe(expectedState);
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

  describe('noChanges()', () => {
    it('reports if there has been any changes to the form', () => {
      component.ngOnInit();
      expect(component.noChanges()).toEqual(true);
      component.toggleInput(component.eventFeedRemoveData, true);
      expect(component.noChanges()).toEqual(false);
    });
  });

  describe('when jobSchedulerStatus is set', () => {
  //   beforeAll(() => {
  //     const jobMissingNodes: IngestJob = {
  //       running: false,
  //       name: IngestJobs.MissingNodes,
  //       threshold: '60m',
  //       every: '1h'
  //     };
  //     const jobMissingNodesForDeletion: IngestJob = {
  //       running: true,
  //       name: IngestJobs.MissingNodesForDeletion,
  //       threshold: '24h',
  //       every: '60m'
  //     };
  //     mockJobSchedulerStatus = new JobSchedulerStatus(true, [
  //       jobMissingNodes,
  //       jobMissingNodesForDeletion
  //     ]);
  //   });

    // it('updates the "missingNodes" form group correctly', () => {
    //   component.updateForm(mockJobSchedulerStatus);

    //   const missingNodesValues = component.automateSettingsForm
    //     .controls.missingNodes.value;

    //   expect(missingNodesValues.disable).toEqual(true);
    //   expect(missingNodesValues.threshold).toEqual('60');
    //   expect(missingNodesValues.unit).toEqual('m');
    // });

    // it('updates the "deleteMissingNodes" form group correctly', () => {
    //   component.updateForm(mockJobSchedulerStatus);

    //   const deleteMssingNodesValues = component.automateSettingsForm
    //     .controls.deleteMissingNodes.value;

    //   expect(deleteMssingNodesValues.disable).toEqual(false);
    //   expect(deleteMssingNodesValues.threshold).toEqual('24');
    //   expect(deleteMssingNodesValues.unit).toEqual('h');
    // });

    // it('does not updates the "eventFeed" form group', () => {
    //   component.updateForm(mockJobSchedulerStatus);

    //   const eventFeedValues = component.automateSettingsForm
    //     .controls.eventFeed.value;

    //   // These are the defaults
    //   expect(eventFeedValues.disable).toEqual(false);
    //   expect(eventFeedValues.threshold).toEqual('');
    //   expect(eventFeedValues.unit).toEqual('d');
    // });

    // describe('when user applyChanges()', () => {
    //   it('saves settings', () => {
    //     component.updateForm(mockJobSchedulerStatus);
    //     component.applyChanges();
    //     expect(component.formChanged).toEqual(false);
    //     expect(component.notificationType).toEqual('info');
    //     expect(component.notificationMessage)
    //       .toEqual('All settings have been updated successfully');
    //     expect(component.notificationVisible).toEqual(true);
    //   });

      // xdescribe('and there is an error', () => {
      //   it('triggers a notification error (shows a banner)', () => {
      //     component.updateForm(mockJobSchedulerStatus);
      //     component.applyChanges();
      //     expect(component.notificationType).toEqual('error');
      //     expect(component.notificationMessage)
      //       .toEqual('Unable to update one or more settings. Verify the console logs.');
      //     expect(component.notificationVisible).toEqual(true);
      //   });
      // });

    });

});

