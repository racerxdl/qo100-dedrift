import {DefinedActions} from "./actions";
import {FFTInitialState, MetricsInitialState, SettingsInitialState, StatusInitialState} from "./initialStates";
import {combineReducers} from "redux";
import {SettingsState, StatusState} from "./types";


function segmentFFT(state: any | void | null, action: any) {
  if (action.type === DefinedActions.AddSegmentFFT) {
    const s = !state ? FFTInitialState : state;
    return {
      ...s,
      samples: action.samples,
      sampleRate: action.sampleRate,
      centerFrequency: action.centerFrequency,
    };
  }

  return state || FFTInitialState;
}

function fullFFT(state: any | void | null, action: any) {
  if (action.type === DefinedActions.AddFullFFT) {
    const s = !state ? FFTInitialState : state;
    return {
      ...s,
      samples: action.samples,
      sampleRate: action.sampleRate,
      centerFrequency: action.centerFrequency,
    };
  }

  return state || FFTInitialState;
}

function metrics(state: any | void | null, action: any) {
  if (action.type === DefinedActions.AddMetrics) {
    const s = !state ? MetricsInitialState : state;
    return {
      ...s,
      metrics: action.metrics,
    };
  }

  return state || MetricsInitialState;
}

function settings(state: any | void | null, action: any) {
  if (action.type === DefinedActions.SetSettings) {
    const s = !state ? SettingsInitialState : state;
    return {
      ...s,
      ...<SettingsState>action,
    }
  }

  return state || SettingsInitialState;
}

function status(state: any | void | null, action: any) {
  if (action.type === DefinedActions.SetStatus) {
    const s = !state ? StatusInitialState : state;
    return {
      ...s,
      ...<StatusState>action,
    }
  }

  return state || StatusInitialState;
}

export default combineReducers({
  segmentFFT,
  fullFFT,
  metrics,
  settings,
  status,
})
