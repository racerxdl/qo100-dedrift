package main

import (
	"encoding/binary"
	"flag"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/config"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"os"
	"os/signal"
	"runtime/pprof"
)

const (
	ConfigFileName = "qo100.toml"
)

func init() {
	slog.SetShowLines(true)
	slog.SetScopeLength(26)
}

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

	InitDSP()

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

	server = rtltcp.MakeRTLTCPServer(":1234")
	server.SetOnCommand(func(sessionId string, cmd rtltcp.Command) {
		client.SendCommand(cmd)
		if cmd.Type == rtltcp.SetFrequency {
			frequency := binary.BigEndian.Uint32(cmd.Param[:])
			OnChangeFrequency(frequency)
		}
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
