package dedrift

import (
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/metrics"
	"github.com/racerxdl/segdsp/dsp"
	"github.com/racerxdl/segdsp/tools"
	"math"
	"time"
)

const (
	TwoPi        = float32(math.Pi * 2)
	MinusTwoPi   = -TwoPi
	OneOverTwoPi = float32(1 / (2 * math.Pi))
)

var fftN = []int{1024, 2048, 4096, 8192, 16384}

func (w *Worker) initDSP() error {
	outSampleRate := float64(w.pc.Source.SampleRate) / float64(w.pc.Processing.WorkDecimation)
	translatorTaps := dsp.MakeLowPass(w.pc.Processing.Translation.Gain, float64(w.pc.Source.SampleRate), (outSampleRate/2)-w.pc.Processing.Translation.TransitionWidth, w.pc.Processing.Translation.TransitionWidth)
	//translatorTaps := dsp.MakeLowPassFixed(pc.Processing.Translation.Gain, float64(pc.Source.SampleRate), outSampleRate/2, 32)
	slog.Info("Translator Taps Length: %d", len(translatorTaps))
	w.translator = dsp.MakeFrequencyTranslator(int(w.pc.Processing.WorkDecimation), -w.pc.Processing.BeaconOffset, float32(w.pc.Source.SampleRate), translatorTaps)
	w.agc = dsp.MakeAttackDecayAGC(w.pc.Processing.AGC.AttackRate, w.pc.Processing.AGC.DecayRate, w.pc.Processing.AGC.Reference, w.pc.Processing.AGC.Gain, w.pc.Processing.AGC.MaxGain)
	w.costas = dsp.MakeCostasLoop2(w.pc.Processing.CostasLoop.Bandwidth)
	w.interp = dsp.MakeFloatInterpolator(int(w.pc.Processing.WorkDecimation))
	slog.Info("Output Sample Rate: %f", outSampleRate)
	w.dcblock = dsp.MakeDCFilter()
	w.fftWindow = dsp.HammingWindow(fftSize)

	metrics.SegmentSampleRate.Set(float64(outSampleRate))
	metrics.SegmentCenterFrequency.Set(float64(w.beaconAbsoluteFrequency))

	return nil
}

func (w *Worker) dsp() {
	log := w.log

	log.Info("Starting DSP Loop")

	w.beaconAbsoluteFrequency = w.pc.Source.CenterFrequency + uint32(w.pc.Processing.BeaconOffset)
	log.Info("Beacon absolute frequency: %d Hz", w.beaconAbsoluteFrequency)

	sampleRate := float32(w.pc.Source.SampleRate)
	segSampleRate := sampleRate / float32(w.pc.Processing.WorkDecimation)

	for w.dspRunning {
		for w.sampleFifo.Len() == 0 {
			time.Sleep(time.Millisecond * 5)
			if !w.dspRunning {
				break
			}
		}

		if !w.dspRunning {
			break
		}

		originalData := w.sampleFifo.Next().([]complex64)
		w.dcblock.WorkInline(originalData)

		w.checkAndResizeBuffers(len(originalData))

		a := w.buffer0
		b := w.buffer1

		a = a[:len(originalData)]

		copy(a, originalData)

		l := w.translator.WorkBuffer(a, b)
		swapAndTrimSlices(&a, &b, l)

		l = w.agc.WorkBuffer(a, b)
		swapAndTrimSlices(&a, &b, l)

		l = w.costas.WorkBuffer(a, b)
		swapAndTrimSlices(&a, &b, l)

		if time.Since(w.lastShiftReport) > time.Second {
			hzDrift := w.costas.GetFrequency() * (float32(w.pc.Source.SampleRate) / float32(w.pc.Processing.WorkDecimation)) / (math.Pi * 2)
			//log.Info("Offset: %f Hz", hzDrift)
			metrics.LockOffset.Set(float64(hzDrift))
			metrics.SegmentCenterFrequency.Set(float64(w.beaconAbsoluteFrequency) + float64(hzDrift))
			w.lastShiftReport = time.Now()
		}

		fs := w.costas.GetFrequencyShift()
		fs = w.interp.Work(fs)

		for i, v := range fs {
			c := tools.PhaseToComplex(w.phase)
			originalData[i] *= c
			w.phase -= v
			if w.phase > TwoPi || w.phase < MinusTwoPi { // Wrap phase between - 2 * pi and + 2 * pi
				w.phase = w.phase*OneOverTwoPi - float32(int(w.phase*OneOverTwoPi))
				w.phase = w.phase * TwoPi
			}
		}

		if w.onData != nil {
			w.onData(originalData)
		}

		if time.Since(w.lastFFT) > fftInterval && w.onFFT != nil {
			var segFFT []float32
			var fullFFT []float32

			if w.pc.Server.WebSettings.HighQualityFFT {
				segFFT = w.computeHQFFT(segSampleRate, a, w.lastSegFFT)
				fullFFT = w.computeHQFFT(sampleRate, originalData, w.lastFullFFT)
			} else {
				segFFT = w.computeFFT(segSampleRate, a, w.lastSegFFT)
				fullFFT = w.computeFFT(sampleRate, originalData, w.lastFullFFT)
			}

			w.onFFT(segFFT, fullFFT)

			w.lastFFT = time.Now()
		}
	}
}

func (w *Worker) OnChangeFrequency(newFrequency uint32) {
	log := w.log
	log.Info("Changed center frequency from %d Hz to %d Hz. Recalculating.", w.pc.Source.CenterFrequency, newFrequency)
	w.pc.Processing.BeaconOffset = float32(w.beaconAbsoluteFrequency - newFrequency)
	w.pc.Source.CenterFrequency = newFrequency
	w.refreshDSP()
}

func (w *Worker) refreshDSP() {
	w.translator.SetFrequency(-w.pc.Processing.BeaconOffset)
	w.log.Debug("New Beacon Offset: %f Hz", w.pc.Processing.BeaconOffset)
	metrics.ServerCenterFrequency.Set(float64(w.pc.Source.CenterFrequency))
}
