package main

import (
	"bufio"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/quan-to/slog"
	"github.com/racerxdl/fastconvert"
	"os"
	"runtime"
	"sync"
	"time"
)

const SampleTypeFloatIQ = 0
const SampleTypeS16IQ = 1
const SampleTypeS8IQ = 2

type SampleCallbackData struct {
	ComplexArray []complex64
	Int16Array   []int16
	Int8Array    []int8
	SampleType   int
	NumSamples   int
}

type SamplesCallback func(data SampleCallbackData)

const CFileFrontendBufferSize = 65535

// region Struct Definition
type CFileFrontend struct {
	sync.Mutex
	running         bool
	filename        string
	cb              SamplesCallback
	sampleRate      uint32
	centerFrequency uint32
	sampleBuffer    []complex64
	fileHandler     *os.File
	t0              time.Time
	fastAsPossible  bool
}

// endregion
// region Constructor
func NewCFileFrontend(filename string) *CFileFrontend {
	return &CFileFrontend{
		filename:        filename,
		running:         false,
		sampleRate:      0,
		centerFrequency: 0,
		cb:              nil,
		fastAsPossible:  false,
	}
}

// endregion
// region Getters
func (f *CFileFrontend) GetName() string {
	return fmt.Sprintf("CFileFrontend (%s)", f.filename)
}
func (f *CFileFrontend) GetShortName() string {
	return "CFileFrontend"
}
func (f *CFileFrontend) GetAvailableSampleRates() []uint32 {
	return make([]uint32, 0)
}
func (f *CFileFrontend) GetCenterFrequency() uint32 {
	return f.centerFrequency
}
func (f *CFileFrontend) GetSampleRate() uint32 {
	return f.sampleRate
}
func (f *CFileFrontend) EnableFastAsPossible() {
	f.fastAsPossible = true
}

// endregion
// region Setters
func (f *CFileFrontend) SetSamplesAvailableCallback(cb SamplesCallback) {
	f.cb = cb
}
func (f *CFileFrontend) SetSampleRate(sampleRate uint32) uint32 {
	f.sampleRate = sampleRate
	return sampleRate
}
func (f *CFileFrontend) SetCenterFrequency(centerFrequency uint32) uint32 {
	f.centerFrequency = centerFrequency
	return centerFrequency
}

// endregion
// region Commands
func (f *CFileFrontend) Init() bool {
	return true
}
func (f *CFileFrontend) Destroy() {}
func (f *CFileFrontend) isRunning() bool {
	f.Lock()
	defer f.Unlock()
	return f.running
}
func (f *CFileFrontend) Start() {
	f.Lock()
	defer f.Unlock()

	if f.running {
		slog.Error("CFileFrontend is already running.")
		return
	}

	f.running = true

	go func(frontend *CFileFrontend) {
		slog.Info("CFileFrontend Routine started")
		f, err := os.Open(f.filename)

		var period = time.Second * CFileFrontendBufferSize / time.Duration(frontend.sampleRate)

		if err != nil {
			slog.Error("Error opening file %s: %s", Bold(frontend.filename), Bold(err))
			frontend.running = false
			return
		}
		defer frontend.fileHandler.Close()

		frontend.fileHandler = f
		frontend.t0 = time.Now()
		frontend.sampleBuffer = make([]complex64, CFileFrontendBufferSize)

		var reader = bufio.NewReader(f)

		if frontend.fastAsPossible {
			period /= 8 // Avoid lock up
		}

		buff := make([]byte, len(frontend.sampleBuffer)*4*2)

		for frontend.isRunning() {
			if time.Since(frontend.t0) >= period {
				_, err = reader.Read(buff)
				if err != nil {
					slog.Error("Error reading input CFile: %s", Bold(err))
					frontend.running = false
					break
				}
				fastconvert.ReadByteArrayToComplex64Array(buff, frontend.sampleBuffer)
				if frontend.cb != nil {
					var cbData = SampleCallbackData{
						SampleType:   SampleTypeFloatIQ,
						NumSamples:   len(frontend.sampleBuffer),
						ComplexArray: frontend.sampleBuffer,
					}
					frontend.cb(cbData)
				}
				frontend.t0 = time.Now()
			}
			runtime.Gosched()
		}
		slog.Error("CFileFrontend Routine ended")
	}(f)
}

func (f *CFileFrontend) Stop() {
	f.Lock()
	defer f.Unlock()

	if !f.running {
		slog.Error("CFileFrontend is not running")
		return
	}

	f.running = false
}

func (f *CFileFrontend) SetAntenna(string) {}
func (f *CFileFrontend) SetAGC(bool)       {}
func (f *CFileFrontend) SetGain1(uint8)    {}
func (f *CFileFrontend) SetGain2(uint8)    {}
func (f *CFileFrontend) SetGain3(uint8)    {}
func (f *CFileFrontend) SetBiasT(bool)     {}

// endregion
