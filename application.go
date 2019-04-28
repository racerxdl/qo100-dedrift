package main

import (
	"flag"
	"github.com/quan-to/slog"
	"github.com/racerxdl/qo100-dedrift/config"
	"github.com/racerxdl/qo100-dedrift/dedrift"
	"github.com/racerxdl/qo100-dedrift/metrics"
	"github.com/racerxdl/qo100-dedrift/rpc"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"github.com/racerxdl/qo100-dedrift/web"
	"os"
	"os/signal"
	"runtime/pprof"
)

func init() {
	slog.SetShowLines(true)
	slog.SetScopeLength(26)
}

var log = slog.Scope("Application")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var createDefault = flag.Bool("defaultConfig", false, "write a default config file")
var pc config.ProgramConfig

func main() {
	var err error
	flag.Parse()
	// region CPU Profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	// endregion
	// region Create Default Configuration
	if *createDefault {
		log.Info("Create Default Config enabled. Saving defaults to %s", config.ConfigFileName)
		err = config.SaveConfig(config.ConfigFileName, config.DefaultConfig)

		if err != nil {
			log.Fatal(err)
		}

		return
	}

	pc, err = config.LoadConfig(config.ConfigFileName)

	if err != nil {
		log.Fatal("Error loading configuration file at %s: %s", config.ConfigFileName, err)
	}
	// endregion
	// region RTLTCP Client Configuration
	client := rtltcp.MakeClient()
	err = client.Connect(pc.Source.Address)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Stop()

	_ = client.SetSampleRate(pc.Source.SampleRate)
	_ = client.SetCenterFrequency(pc.Source.CenterFrequency)
	_ = client.SetGain(uint32(pc.Source.Gain * 10))
	// endregion
	// region Metrics Setup
	metrics.MaxWebConnections.Add(float64(pc.Server.MaxWebConnections))
	metrics.MaxConnections.Add(float64(pc.Server.MaxRTLConnections))
	metrics.ServerCenterFrequency.Set(float64(pc.Source.CenterFrequency))
	metrics.ServerSampleRate.Set(float64(pc.Source.SampleRate))
	// endregion
	// region RTLTCP Server Setup
	server := rtltcp.MakeRTLTCPServer(pc.Server.RTLTCPAddress)
	server.SetDongleInfo(client.GetDongleInfo())

	err = server.Start()
	if err != nil {
		log.Fatal("Error starting RTLTCP: %s", err)
	}

	defer server.Stop()
	// endregion
	// region DSP Initialization
	worker := dedrift.MakeWorker(pc, client, server)
	err = worker.Init()
	if err != nil {
		log.Fatal(err)
	}
	// endregion
	// region Web Server Setup
	ws := web.MakeWebServer(pc.Server.HTTPAddress, pc.Server.MaxWebConnections, pc.Server.WebSettings)
	err = ws.Start()

	if err != nil {
		log.Fatal(err)
	}

	defer ws.Stop()

	worker.SetOnFFT(func(segFFT, fullFFT []float32) {
		ws.BroadcastFFT(web.MessageTypeMainFFT, fullFFT)
		ws.BroadcastFFT(web.MessageTypeSegFFT, segFFT)
	})
	// endregion
	// region RPC
	rpcServer := rpc.MakeRPCServer(pc.Server.RPC.Username, pc.Server.RPC.Password, worker)
	ws.RegisterRPC(rpcServer.RegisterURLs)
	// endregion
	// region Main Loop Setup
	worker.Start()

	var sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig

	worker.Stop()
	// endregion
}
