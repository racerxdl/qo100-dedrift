import {FFTAction, MetricsAction, SettingsAction, SettingsState, StatusAction, StatusState} from "./types";
import {Metric} from "../Tools/types";

const DefinedActions = {
  AddSegmentFFT: 'ADD_SEGMENT_FFT',
  AddFullFFT: 'ADD_FULL_FFT',
  AddMetrics: 'ADD_METRICS',
  SetSettings: 'SET_SETTINGS',
  SetStatus: 'SET_STATUS',
};


function AddMetrics(metrics: Metric[]): MetricsAction {
  return {
    type: DefinedActions.AddMetrics,
    metrics,
  }
}

function AddSegmentFFT(centerFrequency: number, sampleRate: number, samples: number[]): FFTAction {
  return {
    type: DefinedActions.AddSegmentFFT,
    samples,
    centerFrequency,
    sampleRate,
  }
}


function AddFullFFT(centerFrequency: number, sampleRate: number, samples: number[]): FFTAction {
  return {
    type: DefinedActions.AddFullFFT,
    samples,
    centerFrequency,
    sampleRate,
  }
}

function SetSettings(settings: SettingsState): SettingsAction {
  return {
    type: DefinedActions.SetSettings,
    ...settings,
  }
}

function SetStatus(status: StatusState): StatusAction {
  return {
    type: DefinedActions.SetStatus,
    ...status,
  }
}

export {
  DefinedActions,
  AddMetrics,
  AddFullFFT,
  AddSegmentFFT,
  SetSettings,
  SetStatus,
}
