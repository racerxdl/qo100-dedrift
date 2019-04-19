import {FFTState, MetricsState, SettingsState, StatusState} from './types';

const FFTInitialState: FFTState = {
  samples: [],
  centerFrequency: 0,
  sampleRate: 0,
};

const MetricsInitialState: MetricsState = {
  metrics: []
};

const SettingsInitialState: SettingsState = {
  name: "PU2NVX Test Server",
  segFFT: {
    maxVal: -70,
    range: 40,
    width: 512,
    height: 256
  },
  fullFFT: {
    maxVal: -70,
    range: 40,
    width: 512,
    height: 256
  }
};

const StatusInitialState: StatusState = {
  wsConnected: false,
};

export {
  FFTInitialState,
  MetricsInitialState,
  SettingsInitialState,
  StatusInitialState,
}
