package main

import (
	"bytes"
	"encoding/binary"
	"github.com/opensatelliteproject/SatHelperApp/Frontend"
	"github.com/quan-to/slog"
	"github.com/racerxdl/go.fifo"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"github.com/racerxdl/segdsp/dsp"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
	"unsafe"
)

var log = slog.Scope("Application")

var conns []net.Conn

var connMtx = sync.Mutex{}

const defaultReadTimeout = 1000

var dongleInfo = rtltcp.DongleInfo{
	Magic:          [4]uint8{'R', 'T', 'L', '0'},
	TunerType:      rtltcp.RtlsdrTunerR820t,
	TunerGainCount: 4,
}

func handlePacket(clog *slog.Instance, cmd rtltcp.Command) {
	uParam := binary.BigEndian.Uint32(cmd.Param[:])
	clog.Debug("Received Type %s (%d) with arg (%d) %v", rtltcp.CommandTypeToName[cmd.Type], cmd.Type, uParam, cmd.Param)
	//switch cmd.Type {
	//    case rtltcp.SetFrequency:
	//
	//}
}

func handleRequest(conn net.Conn) {
	clog := slog.Scope(conn.RemoteAddr().String())
	clog.Info("Received connection")

	clog.Debug("Sending greeting with DongleInfo")
	err := binary.Write(conn, binary.LittleEndian, &dongleInfo)
	if err != nil {
		clog.Error("Error sending greeting: %s", err)
		return
	}

	// Adding to connection pool
	connMtx.Lock()
	conns = append(conns, conn)
	connMtx.Unlock()

	cmd := rtltcp.Command{}
	buffer := make([]byte, unsafe.Sizeof(cmd))
	running := true

	for running {
		_ = conn.SetReadDeadline(time.Now().Add(defaultReadTimeout))
		n, err := conn.Read(buffer)

		if err != nil {
			if err.Error() != "EOF" {
				switch e := err.(type) {
				case net.Error:
					if !e.Timeout() {
						if running {
							clog.Error("Error receiving data: %s", e)
						}
						running = false
					}
				default:
					clog.Error("Error receiving data: %s", e)
					running = false
				}
			} else {
				running = false
			}
		}

		if !running {
			break
		}

		if n > 0 {
			b := bytes.NewReader(buffer)
			err := binary.Read(b, binary.LittleEndian, &cmd)
			if err != nil {
				clog.Error("Error parsing packet: %s", err)
				continue
			}
			handlePacket(clog, cmd)
		}

	}
	connMtx.Lock()
	for i, v := range conns {
		if conn.RemoteAddr().String() == v.RemoteAddr().String() {
			conns = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	connMtx.Unlock()
	_ = conn.Close()

	clog.Info("Connection closed.")
}

func PostSamples(data []complex64) {
	iqBytes := make([]byte, len(data)*2)

	for i, v := range data {
		rv := 256 - real(v)*127
		iv := 256 - imag(v)*127

		iqBytes[i*2] = uint8(rv)
		iqBytes[i*2+1] = uint8(iv)
	}

	connMtx.Lock()
	for _, v := range conns {
		_, _ = v.Write(iqBytes)
	}
	connMtx.Unlock()
}

func Server() {
	conns = make([]net.Conn, 0)
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		slog.Fatal("Error listening:", err.Error())
	}
	defer l.Close()
	slog.Info("Listening on :1234")
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			slog.Fatal("Error accepting: ", err.Error())
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

var buffer0 []complex64
var buffer1 []complex64

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

var sampleFifo = fifo.NewQueue()
var dspRunning bool

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

		//checkAndResizeBuffers(len(originalData))
		//
		//a := buffer0
		//b := buffer1
		//
		//copy(a, originalData)
		//
		//l := preFilter.WorkBuffer(a, b)
		//swapAndTrimSlices(&a, &b, l)
		//
		//l = translator.WorkBuffer(a, b)
		//swapAndTrimSlices(&a, &b, l)
		//
		//l = postFilter.WorkBuffer(a, b)
		//swapAndTrimSlices(&a, &b, l)
		//
		//l = agc.WorkBuffer(a, b)
		//swapAndTrimSlices(&a, &b, l)
		//
		//l = costas.WorkBuffer(a, b)
		//swapAndTrimSlices(&a, &b, l)

		o := originalData
		//o = preFilter.Work(o)
		//o = translator.Work(o)
		//o = postFilter.Work(o)
		//o = agc.Work(o)
		//o = costas.Work(o)

		slog.Info("Offset: %f", costas.GetFrequency())

		//c := complex(tools.Cos(costas.GetFrequency()), tools.Sin(costas.GetFrequency()))
		//
		//rotator.SetPhaseIncrement(c)
		//
		//originalData = rotator.Work(originalData)

		PostSamples(o)
	}
}

var preFilter *dsp.FirFilter
var translator *dsp.FrequencyTranslator
var postFilter *dsp.FirFilter
var agc *dsp.AttackDecayAGC
var costas dsp.CostasLoop
var rotator *dsp.Rotator

var shiftTaps = []float32{
	0.001253673923201859, 0.007188489194959402, 0.007526693399995565, -0.0008259809110313654, -0.013517701998353004,
	-0.01700032874941826, -0.0008620411390438676, 0.02520219422876835, 0.03342488035559654, 0.006000528112053871,
	-0.03816521540284157, -0.053771745413541794, -0.015265795402228832, 0.04817265644669533, 0.07347673922777176,
	0.027454184368252754, -0.052131328731775284, -0.08773830533027649, -0.039813026785850525, 0.04896785691380501,
	0.09293936938047409, 0.04896785691380501, -0.039813026785850525, -0.08773830533027649, -0.052131328731775284,
	0.027454184368252754, 0.07347673922777176, 0.04817265644669533, -0.015265795402228832, -0.053771745413541794,
	-0.03816521540284157, 0.006000528112053871, 0.03342488035559654, 0.02520219422876835, -0.0008620411390438676,
	-0.01700032874941826, -0.013517701998353004, -0.0008259809110313654, 0.007526693399995565, 0.007188489194959402,
	0.001253673923201859,
}

const (
	//SampleRate = 900e3
	SampleRate     = 56250
	BeaconOffset   = 140e3
	WorkDecimation = 16
	//inputIQ = "/media/ELTN/HAM/baseband/QO-100/SDRSharp_20190311_011013Z_739764000Hz_IQ_900e3.cfile"
	inputIQ = "/media/ELTN/HAM/baseband/QO-100/SDRSharp_20190311_011013Z_739764000Hz_IQ_56250-N.cfile"
)

func main() {
	preFilter = dsp.MakeFirFilter(shiftTaps)
	translatorLowPass := dsp.MakeLowPass(1, SampleRate, SampleRate/2-60e3, 10e3)
	translator = dsp.MakeFrequencyTranslator(1, BeaconOffset, SampleRate, translatorLowPass)
	postTaps := dsp.MakeLowPass(1, SampleRate/WorkDecimation, 20e3, 5e3)
	postFilter = dsp.MakeDecimationFirFilter(16, postTaps)
	agc = dsp.MakeAttackDecayAGC(1e-2, 0.2, 1, 10, 65536)
	costas = dsp.MakeCostasLoop2(10e-3)
	rotator = dsp.MakeRotator()

	src := NewCFileFrontend(inputIQ)
	src.SetSampleRate(SampleRate)
	src.SetSamplesAvailableCallback(func(data Frontend.SampleCallbackData) {
		sampleFifo.Add(data.ComplexArray)
	})

	slog.Info("Output Sample Rate: %f", SampleRate/WorkDecimation)

	dspRunning = true
	go DSP()
	go Server()
	src.Start()

	var sig chan os.Signal
	sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	dspRunning = false
}
