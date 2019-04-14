package main

import (
	"flag"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/config"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"github.com/racerxdl/segdsp/dsp"
	"os"
	"os/signal"
	"runtime/pprof"
)

const (
	ConfigFileName = "qo100.toml"
	//SampleRate     = 1800e3
	//CenterFrequency = 739810000
	//BeaconOffset   = 125e3
	//CenterFrequency = 740000000
	//BeaconOffset    = 140e3
	//WorkDecimation  = 32
	//rtlServer = "187.85.15.201:1234"
	//rtlServer = "127.0.0.1:1235"
)

var log = slog.Scope("Application")
var server *rtltcp.Server
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var createDefault = flag.Bool("defaultConfig", false, "write a default config file")
var pc config.ProgramConfig

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

	if *createDefault {
		slog.Info("Create Default Config enabled. Saving defaults to %s", ConfigFileName)
		err = config.SaveConfig(ConfigFileName, config.DefaultConfig)

		if err != nil {
			slog.Fatal(err)
		}

		return
	}

	pc, err = config.LoadConfig(ConfigFileName)

	if err != nil {
		slog.Fatal("Error loading configuration file at %s: %s", ConfigFileName, err)
	}

	outSampleRate := float64(pc.Source.SampleRate) / float64(pc.Processing.WorkDecimation)
	translatorTaps := dsp.MakeLowPass(pc.Processing.Translation.Gain, float64(pc.Source.SampleRate), (outSampleRate/2)-pc.Processing.Translation.TransitionWidth, pc.Processing.Translation.TransitionWidth)
	slog.Info("Translator Taps Length: %d", len(translatorTaps))
	translator = dsp.MakeFrequencyTranslator(int(pc.Processing.WorkDecimation), -pc.Processing.BeaconOffset, float32(pc.Source.SampleRate), translatorTaps)
	agc = dsp.MakeAttackDecayAGC(pc.Processing.AGC.AttackRate, pc.Processing.AGC.DecayRate, pc.Processing.AGC.Reference, pc.Processing.AGC.Gain, pc.Processing.AGC.MaxGain)
	costas = dsp.MakeCostasLoop2(pc.Processing.CostasLoop.Bandwidth)
	interp = dsp.MakeFloatInterpolator(int(pc.Processing.WorkDecimation))

	client := rtltcp.MakeClient()
	err = client.Connect(pc.Source.Address)
	if err != nil {
		slog.Fatal(err)
	}

	client.SetSampleRate(pc.Source.SampleRate)
	client.SetCenterFrequency(pc.Source.CenterFrequency)
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
