package main

import (
	"github.com/racerxdl/segdsp/dsp"
	"github.com/racerxdl/segdsp/dsp/fft"
	"github.com/racerxdl/segdsp/tools"
	"math"
	"time"
)

const (
	FFTFPS       = 15
	fftInterval  = time.Second / FFTFPS
	fftSize      = 1024
	FFTAveraging = 2
)

var lastFFT = time.Now()
var onFFT func(segFFT, fullFFT []float32)
var fftWindow []float64

var lastFullFFT = make([]float32, fftSize)
var lastSegFFT = make([]float32, fftSize)

func ComputeFFT(sampleRate float32, samples []complex64, lastFFT []float32) []float32 {
	samples = samples[:fftSize]

	if len(fftWindow) != fftSize {
		fftWindow = dsp.HammingWindow(fftSize)
	}

	// Apply window to samples
	for j := 0; j < len(samples); j++ {
		var s = samples[j]
		var r = real(s) * float32(fftWindow[j])
		var i = imag(s) * float32(fftWindow[j])
		samples[j] = complex(r, i)
	}

	fftCData := fft.FFT(samples)

	var fftSamples = make([]float32, len(fftCData))
	var l = len(fftSamples)
	var lastV = float32(0)
	for i, v := range fftCData {
		var oI = (i + l/2) % l
		var m = float64(tools.ComplexAbsSquared(v) * (1.0 / sampleRate))

		m = 10 * math.Log10(m)

		fftSamples[oI] = (lastFFT[i]*(FFTAveraging-1) + float32(m)) / FFTAveraging
		if fftSamples[i] != fftSamples[i] { // IsNaN
			fftSamples[i] = 0
		}

		if i > 0 {
			fftSamples[oI] = lastV*0.4 + fftSamples[oI]*0.6
		}

		lastV = fftSamples[oI]
	}

	copy(lastFFT, fftSamples)

	return fftSamples
}

func ComputeHQFFT(sampleRate float32, samples []complex64, lastFFT []float32) []float32 {
	hqFFTLength := int(fftSize)

	for _, v := range fftN {
		if v > hqFFTLength && v <= len(samples) {
			hqFFTLength = v
		} else {
			break
		}
	}

	nDiv := hqFFTLength / fftSize

	samples = samples[:hqFFTLength]

	if len(fftWindow) != hqFFTLength {
		fftWindow = dsp.HammingWindow(hqFFTLength)
	}

	// Apply window to samples
	for j := 0; j < len(samples); j++ {
		var s = samples[j]
		var r = real(s) * float32(fftWindow[j])
		var i = imag(s) * float32(fftWindow[j])
		samples[j] = complex(r, i)
	}

	fftCData := fft.FFT(samples)

	var fftSamples = make([]float32, fftSize)
	var lastV = float32(0)

	for i := 0; i < fftSize; i++ {
		// Compute Average of nDivs
		v := float32(0)

		for _, v2 := range fftCData[i*nDiv : (i+1)*nDiv] {
			var m = float64(tools.ComplexAbsSquared(v2) * (1.0 / sampleRate))
			v += float32(10 * math.Log10(m))
		}

		v /= float32(nDiv)

		// Put in the right output place
		var oI = (i + fftSize/2) % fftSize

		fftSamples[oI] = (lastFFT[i]*(FFTAveraging-1) + float32(v)) / FFTAveraging
		if fftSamples[i] != fftSamples[i] { // IsNaN
			fftSamples[i] = 0
		}

		if i > 0 {
			fftSamples[oI] = lastV*0.4 + fftSamples[oI]*0.6
		}

		lastV = fftSamples[oI]
	}

	copy(lastFFT, fftSamples)

	return fftSamples
}

func SetOnFFT(cb func(segFFT, fullFFT []float32)) {
	onFFT = cb
}
