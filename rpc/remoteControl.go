package rpc

import "github.com/racerxdl/qo100-dedrift/config"

type RemoteControl interface {
	SetGain(gain uint32) error
	SetSampleRate(sampleRate uint32) error
	SetCenterFrequency(centerFrequency uint32) error
	SetBeaconOffset(offset float32) error
	SetSegFFTConfig(setting config.FFTWindowSetting) error
	SetFullFFTConfig(setting config.FFTWindowSetting) error
	Save() error
}
