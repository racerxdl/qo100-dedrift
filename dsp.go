package main

import (
	"github.com/quan-to/slog"
	"github.com/racerxdl/go.fifo"
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

var buffer0 []complex64
var buffer1 []complex64

var translator *dsp.FrequencyTranslator
var agc *dsp.AttackDecayAGC
var costas dsp.CostasLoop
var interp *dsp.FloatInterpolator

var sampleFifo = fifo.NewQueue()
var dspRunning bool
var lastShiftReport = time.Now()
var phase = float32(0)
var beaconAbsoluteFrequency = uint32(0)

func checkAndResizeBuffers(length int) {
	if len(buffer0) < length {
		buffer0 = make([]complex64, length)
	}
	if len(buffer1) < length {
		buffer1 = make([]complex64, length)
	}
}

func swapAndTrimSlices(a *[]complex64, b *[]complex64, length int) {
	*a = (*a)[:length]
	*b = (*b)[:length]

	c := *b
	*b = *a
	*a = c
}

func OnChangeFrequency(newFrequency uint32) {
	log.Info("Changed center frequency from %d Hz to %d Hz. Recalculating.", pc.Source.CenterFrequency, newFrequency)
	pc.Processing.BeaconOffset = float32(beaconAbsoluteFrequency - newFrequency)
	pc.Source.CenterFrequency = newFrequency
	translator.SetFrequency(-pc.Processing.BeaconOffset)
	log.Debug("New Beacon Offset: %f Hz", pc.Processing.BeaconOffset)
	metrics.ServerCenterFrequency.Set(float64(newFrequency))
}

func InitDSP() {
	outSampleRate := float64(pc.Source.SampleRate) / float64(pc.Processing.WorkDecimation)
	translatorTaps := dsp.MakeLowPass(pc.Processing.Translation.Gain, float64(pc.Source.SampleRate), (outSampleRate/2)-pc.Processing.Translation.TransitionWidth, pc.Processing.Translation.TransitionWidth)
	//translatorTaps := dsp.MakeLowPassFixed(pc.Processing.Translation.Gain, float64(pc.Source.SampleRate), outSampleRate/2, 32)
	slog.Info("Translator Taps Length: %d", len(translatorTaps))
	translator = dsp.MakeFrequencyTranslator(int(pc.Processing.WorkDecimation), -pc.Processing.BeaconOffset, float32(pc.Source.SampleRate), translatorTaps)
	agc = dsp.MakeAttackDecayAGC(pc.Processing.AGC.AttackRate, pc.Processing.AGC.DecayRate, pc.Processing.AGC.Reference, pc.Processing.AGC.Gain, pc.Processing.AGC.MaxGain)
	costas = dsp.MakeCostasLoop2(pc.Processing.CostasLoop.Bandwidth)
	interp = dsp.MakeFloatInterpolator(int(pc.Processing.WorkDecimation))
	slog.Info("Output Sample Rate: %f", outSampleRate)
}

func DSP() {
	log.Info("Starting DSP Loop")

	beaconAbsoluteFrequency = pc.Source.CenterFrequency + uint32(pc.Processing.BeaconOffset)
	log.Info("Beacon absolute frequency: %d Hz", beaconAbsoluteFrequency)

	for dspRunning {
		for sampleFifo.Len() == 0 {
			time.Sleep(time.Millisecond * 5)
			if !dspRunning {
				break
			}
		}

		if !dspRunning {
			break
		}

		originalData := sampleFifo.Next().([]complex64)

		checkAndResizeBuffers(len(originalData))

		a := buffer0
		b := buffer1

		a = a[:len(originalData)]

		copy(a, originalData)

		l := translator.WorkBuffer(a, b)
		swapAndTrimSlices(&a, &b, l)

		l = agc.WorkBuffer(a, b)
		swapAndTrimSlices(&a, &b, l)

		l = costas.WorkBuffer(a, b)
		swapAndTrimSlices(&a, &b, l)

		if time.Since(lastShiftReport) > time.Second {
			hzDrift := costas.GetFrequency() * (float32(pc.Source.SampleRate) / float32(pc.Processing.WorkDecimation)) / (math.Pi * 2)
			//slog.Info("Offset: %f Hz", hzDrift)
			metrics.LockOffset.Set(float64(hzDrift))
			lastShiftReport = time.Now()
		}

		fs := costas.GetFrequencyShift()
		fs = interp.Work(fs)

		for i, v := range fs {
			c := tools.PhaseToComplex(phase)
			originalData[i] *= c
			phase -= v
			if phase > TwoPi || phase < MinusTwoPi {
				phase = phase*OneOverTwoPi - float32(int(phase*OneOverTwoPi))
				phase = phase * TwoPi
			}
		}

		server.ComplexBroadcast(originalData)
	}
}
