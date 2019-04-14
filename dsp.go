package main

import (
	"github.com/quan-to/slog"
	"github.com/racerxdl/go.fifo"
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

func DSP() {
	log.Info("Starting DSP Loop")

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
		//
		l = costas.WorkBuffer(a, b)
		swapAndTrimSlices(&a, &b, l)
		//
		if time.Since(lastShiftReport) > time.Second {
			hzDrift := costas.GetFrequency() * (SampleRate / WorkDecimation) / (math.Pi * 2)
			slog.Info("Offset: %f Hz", hzDrift)
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
