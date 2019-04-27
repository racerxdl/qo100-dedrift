package dedrift

import "github.com/racerxdl/qo100-dedrift/config"

func (w *Worker) SetGain(gain uint32) error {
	w.log.Info("Setting gain to %d", gain)
	return w.client.SetGain(gain)
}

func (w *Worker) SetSampleRate(sampleRate uint32) error {
	w.log.Info("Setting sample rate to %d", sampleRate)
	return w.client.SetSampleRate(sampleRate)
}

func (w *Worker) SetCenterFrequency(centerFrequency uint32) error {
	w.log.Info("Setting center frequency to %d", centerFrequency)
	err := w.client.SetCenterFrequency(centerFrequency)
	if err != nil {
		return err
	}
	w.OnChangeFrequency(centerFrequency)
	return nil
}

func (w *Worker) SetBeaconOffset(offset float32) error {
	w.log.Info("Setting Beacon offset to %f", offset)
	w.pc.Processing.BeaconOffset = offset
	w.refreshDSP()

	return nil
}

func (w *Worker) SetSegFFTConfig(setting config.FFTWindowSetting) error {
	w.log.Info("Changing Beacon FFT Configuration")
	w.pc.Server.WebSettings.SegFFT = setting
	return nil
}

func (w *Worker) SetFullFFTConfig(setting config.FFTWindowSetting) error {
	w.log.Info("Changing Full Spectrum Configuration")
	w.pc.Server.WebSettings.FullFFT = setting
	return nil
}

func (w *Worker) Save() error {
	w.log.Info("Saving configuration")
	return config.SaveConfig(config.ConfigFileName, config.DefaultConfig)
}
