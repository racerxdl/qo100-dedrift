import {Metric} from "../Tools/types";

export type ActionType = {
  type: string
}

export type FFTState = {
  samples: number[]
  centerFrequency: number,
  sampleRate: number,
}

export type MetricsState = {
  metrics: Metric[],
}

export type FFTConfig = {
  maxVal: number;
  range: number;
  width: number;
  height: number;
}

export type SettingsState = {
  name: string;
  segFFT: FFTConfig;
  fullFFT: FFTConfig;
}

export type StatusState = {
  wsConnected: boolean;
}

export type FFTAction = ActionType & FFTState
export type MetricsAction = ActionType & MetricsState;
export type SettingsAction = ActionType & SettingsState;
export type StatusAction = ActionType & StatusState;
