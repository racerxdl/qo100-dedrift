package main

import (
	"flag"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"github.com/racerxdl/segdsp/dsp"
	"os"
	"os/signal"
	"runtime/pprof"
)

const (
	SampleRate = 1800e3
	//CenterFrequency = 739810000
	//BeaconOffset   = 125e3
	CenterFrequency = 740000000
	BeaconOffset    = 140e3
	WorkDecimation  = 32
	//rtlServer = "187.85.15.201:1234"
	rtlServer = "127.0.0.1:1235"
)

var log = slog.Scope("Application")
var server *rtltcp.Server
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	var err error
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	outSampleRate := SampleRate / WorkDecimation
	translatorTaps := dsp.MakeLowPass(64, SampleRate, outSampleRate/2-15e3, 15e3)
	slog.Info("Translator Taps Length: %d", len(translatorTaps))
	translator = dsp.MakeFrequencyTranslator(WorkDecimation, -BeaconOffset, SampleRate, translatorTaps)
	agc = dsp.MakeAttackDecayAGC(0.01, 0.2, 1, 10, 65536)
	costas = dsp.MakeCostasLoop2(0.01)
	interp = dsp.MakeFloatInterpolator(WorkDecimation)

	client := rtltcp.MakeClient()
	err = client.Connect(rtlServer)
	if err != nil {
		slog.Fatal(err)
	}

	client.SetSampleRate(SampleRate)
	client.SetCenterFrequency(CenterFrequency)
	client.SetOnSamples(func(data []complex64) {
		sampleFifo.Add(data)
	})

	slog.Info("Output Sample Rate: %f", outSampleRate)

	server = rtltcp.MakeRTLTCPServer(":1234")
	server.SetOnCommand(func(sessionId string, cmd rtltcp.Command) {
		client.SendCommand(cmd)
	})

	err = server.Start()
	if err != nil {
		slog.Fatal("Error starting RTLTCP: %s", err)
	}

	dspRunning = true
	go DSP()

	var sig chan os.Signal
	sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	running := true

	for running {
		select {
		case <-sig:
			log.Info("Received SIGINT")
			running = false
		}
	}

	client.Stop()
	server.Stop()
	dspRunning = false
}
