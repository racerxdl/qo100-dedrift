package dedrift

import (
	"encoding/binary"
	"github.com/quan-to/slog"
	"github.com/racerxdl/go.fifo"
	"github.com/racerxdl/qo100-dedrift/config"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"github.com/racerxdl/segdsp/dsp"
	"time"
)

type Worker struct {
	pc                      config.ProgramConfig
	translator              *dsp.FrequencyTranslator
	agc                     *dsp.AttackDecayAGC
	costas                  dsp.CostasLoop
	interp                  *dsp.FloatInterpolator
	dcblock                 *dsp.DCFilter
	sampleFifo              *fifo.Queue
	dspRunning              bool
	lastShiftReport         time.Time
	phase                   float32
	beaconAbsoluteFrequency uint32

	buffer0     []complex64
	buffer1     []complex64
	lastFFT     time.Time
	onFFT       OnFFTCallback
	onData      OnIQData
	fftWindow   []float64
	lastFullFFT []float32
	lastSegFFT  []float32

	log    *slog.Instance
	client *rtltcp.Client
	server *rtltcp.Server
}

func MakeWorker(pc config.ProgramConfig, client *rtltcp.Client, server *rtltcp.Server) *Worker {
	w := &Worker{}

	w.sampleFifo = fifo.NewQueue()
	w.dspRunning = false
	w.lastShiftReport = time.Now()
	w.phase = float32(0)
	w.beaconAbsoluteFrequency = uint32(0)

	w.lastFFT = time.Now()
	w.lastFullFFT = make([]float32, fftSize)
	w.lastSegFFT = make([]float32, fftSize)
	w.log = slog.Scope("Worker")
	w.pc = pc
	w.client = client
	w.server = server

	client.SetOnSamples(w.OnSamples)
	server.SetOnCommand(func(sessionId string, cmd rtltcp.Command) bool {
		if cmd.Type == rtltcp.SetSampleRate {
			sampleRate := binary.BigEndian.Uint32(cmd.Param[:])
			if sampleRate != pc.Source.SampleRate {
				w.log.Error("Client asked for %d as sampleRate, but we cannot change it! Current: %d", sampleRate, pc.Source.SampleRate)
				w.log.Error("Closing connection with %s", sessionId)
				return false
			}
			return true
		}

		if pc.Server.AllowControl {
			_ = client.SendCommand(cmd)
			if cmd.Type == rtltcp.SetFrequency {
				frequency := binary.BigEndian.Uint32(cmd.Param[:])
				w.OnChangeFrequency(frequency)
			}
		} else {
			w.log.Warn("Ignoring command %s because AllowControl is false", rtltcp.CommandTypeToName[cmd.Type])
		}

		return true
	})

	return w
}

func (w *Worker) Init() error {
	err := w.initDSP()
	if err != nil {
		return err
	}

	return nil
}

func (w *Worker) SetOnData(cb OnIQData) {
	w.onData = cb
}

func (w *Worker) OnSamples(samples []complex64) {
	w.sampleFifo.Add(samples)
}

func (w *Worker) Start() {
	w.dspRunning = true
	go w.dsp()
}

func (w *Worker) Stop() {
	w.dspRunning = false
}
