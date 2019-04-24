package main

import (
	"encoding/binary"
	"flag"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/config"
	"github.com/racerxdl/qo100-dedrift/metrics"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"github.com/racerxdl/qo100-dedrift/web"
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
		log.Info("Create Default Config enabled. Saving defaults to %s", ConfigFileName)
		err = config.SaveConfig(ConfigFileName, config.DefaultConfig)

		if err != nil {
			log.Fatal(err)
		}

		return
	}

	pc, err = config.LoadConfig(ConfigFileName)

	if err != nil {
		log.Fatal("Error loading configuration file at %s: %s", ConfigFileName, err)
	}

	InitDSP()

	client := rtltcp.MakeClient()
	err = client.Connect(pc.Source.Address)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Stop()

	metrics.MaxWebConnections.Add(float64(pc.Server.MaxWebConnections))
	metrics.MaxConnections.Add(float64(pc.Server.MaxRTLConnections))
	metrics.ServerCenterFrequency.Set(float64(pc.Source.CenterFrequency))
	metrics.ServerSampleRate.Set(float64(pc.Source.SampleRate))

	_ = client.SetSampleRate(pc.Source.SampleRate)
	_ = client.SetCenterFrequency(pc.Source.CenterFrequency)
	client.SetOnSamples(func(data []complex64) {
		sampleFifo.Add(data)
	})
	_ = client.SetGain(uint32(pc.Source.Gain * 10))

	server = rtltcp.MakeRTLTCPServer(pc.Server.RTLTCPAddress)
	server.SetDongleInfo(client.GetDongleInfo())
	server.SetOnCommand(func(sessionId string, cmd rtltcp.Command) bool {
		if cmd.Type == rtltcp.SetSampleRate {
			sampleRate := binary.BigEndian.Uint32(cmd.Param[:])
			if sampleRate != pc.Source.SampleRate {
				log.Error("Client asked for %d as sampleRate, but we cannot change it! Current: %d", sampleRate, pc.Source.SampleRate)
				log.Error("Closing connection with %s", sessionId)
				return false
			}
			return true
		}

		if pc.Server.AllowControl {
			_ = client.SendCommand(cmd)
			if cmd.Type == rtltcp.SetFrequency {
				frequency := binary.BigEndian.Uint32(cmd.Param[:])
				OnChangeFrequency(frequency)
			}
		} else {
			log.Warn("Ignoring command %s because AllowControl is false", rtltcp.CommandTypeToName[cmd.Type])
		}

		return true
	})

	err = server.Start()
	if err != nil {
		log.Fatal("Error starting RTLTCP: %s", err)
	}

	defer server.Stop()

	ws := web.MakeWebServer(pc.Server.HTTPAddress, pc.Server.MaxWebConnections, pc.Server.WebSettings)
	err = ws.Start()

	if err != nil {
		log.Fatal(err)
	}

	defer ws.Stop()

	SetOnFFT(func(segFFT, fullFFT []float32) {
		ws.BroadcastFFT(web.MessageTypeMainFFT, fullFFT)
		ws.BroadcastFFT(web.MessageTypeSegFFT, segFFT)
	})

	dspRunning = true
	go DSP()

	var sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	//running := true

	<-sig

	//for running {
	//	//select {
	//	case <-sig:
	//		log.Info("Received SIGINT")
	//		running = false
	//	//}
	//}
	dspRunning = false
}
