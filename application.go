package main

import (
	"bytes"
	"encoding/binary"
	"github.com/quan-to/slog"
	"github.com/racerxdl/go.fifo"
	"github.com/racerxdl/qo100-dedrift/rtltcp"
	"github.com/racerxdl/segdsp/dsp"
	"github.com/racerxdl/segdsp/tools"
	"math"
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
	err := binary.Write(conn, binary.BigEndian, &dongleInfo)
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
		rv := 128 + real(v)*127
		iv := 128 + imag(v)*127

		if rv < 0 {
			rv = 0
		}
		if rv > 255 {
			rv = 255
		}

		if iv < 0 {
			iv = 0
		}

		if iv > 255 {
			iv = 255
		}

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
var lastShiftReport = time.Now()
var lastFFTUpdate = time.Now()
var phase = float32(0)

const TwoPi = float32(math.Pi * 2)
const MinusTwoPi = -TwoPi
const OneOverTwoPi = float32(1 / (2 * math.Pi))

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
			//partialFFT.SetCenterFrequency(float64(BeaconOffset + hzDrift))
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

		PostSamples(originalData)
		if time.Since(lastFFTUpdate) > time.Second/60 {
			//fullFFT.UpdateSamples(originalData)
			//partialFFT.UpdateSamples(a)
			lastFFTUpdate = time.Now()
		}
	}
}

var translator *dsp.FrequencyTranslator
var agc *dsp.AttackDecayAGC
var costas dsp.CostasLoop
var interp *dsp.FloatInterpolator

//var shiftTaps = []float32{
//	-0.006936056073755026, -0.0059114787727594376, -0.004766103345900774, -0.0035124800633639097, -0.0021645394153892994, -0.0007377190049737692, 0.0007508968701586127, 0.002282458357512951, 0.003836256917566061, 0.005389681551605463, 0.006918228697031736, 0.008395591750741005, 0.009793835692107677, 0.011083660647273064, 0.012234771624207497, 0.013216350227594376, 0.01399762649089098, 0.01454854290932417, 0.014840507879853249, 0.0148472273722291, 0.014545565471053123, 0.013916469179093838, 0.012945862486958504, 0.011625551618635654, 0.009954052045941353, 0.007937340065836906, 0.005589496344327927, 0.002933190669864416, -1.6097297025390217e-17, -0.0031694755889475346, -0.006525722797960043, -0.010010662488639355, -0.01355825737118721, -0.01709536649286747, -0.020542873069643974, -0.023817049339413643, -0.02683117426931858, -0.029497329145669937, -0.03172841668128967, -0.03344028815627098, -0.03455398976802826, -0.034998029470443726, -0.03471070155501366, -0.0336422435939312, -0.031756967306137085, -0.029035158455371857, -0.025474701076745987, -0.021092459559440613, -0.015925239771604538, -0.010030388832092285, -0.0034859066363424063, 0.003609905019402504, 0.011139304377138615, 0.018966492265462875, 0.02693951316177845, 0.03489265590906143, 0.04264933243393898, 0.05002541095018387, 0.056832920759916306, 0.06288408488035202, 0.0679955929517746, 0.07199306041002274, 0.07471556216478348, 0.07602003961801529, 0.07578572630882263, 0.0739181637763977, 0.07035301625728607, 0.06505925953388214, 0.058041974902153015, 0.04934436455368996, 0.039049070328474045, 0.027278704568743706, 0.014195486903190613, -5.148713261814661e-17, -0.01507099810987711, -0.030747657641768456, -0.046731047332286835, -0.06269800662994385, -0.07830678671598434, -0.09320346266031265, -0.10702886432409286, -0.11942609399557114, -0.13004827499389648, -0.13856664299964905, -0.14467853307724, -0.1481153666973114, -0.14865010976791382, -0.14610455930233002, -0.14035560190677643, -0.13134092092514038, -0.11906352639198303, -0.10359520465135574, -0.08507874608039856, -0.06372867524623871, -0.039830684661865234, -0.013739367946982384, 0.014125424437224865, 0.04328383877873421, 0.07320275157690048, 0.1033039465546608, 0.1329735815525055, 0.16157282888889313, 0.1884496808052063, 0.21295134723186493, 0.23443755507469177, 0.25229406356811523, 0.26594626903533936, 0.27487292885780334, 0.2786192297935486, 0.27680930495262146, 0.26915788650512695, 0.2554805874824524, 0.23570281267166138, 0.20986703038215637, 0.17813824117183685, 0.14080704748630524, 0.0982910767197609, 0.051133569329977036, -9.274395350058849e-17, -0.05432789772748947, -0.11096034944057465, -0.16891008615493774, -0.2271050363779068, -0.284402996301651, -0.33960819244384766, -0.3914893567562103, -0.43879905343055725, -0.48029395937919617, -0.5147561430931091, -0.5410138964653015, -0.5579633712768555, -0.5645893216133118, -0.5599846243858337, -0.5433695316314697, -0.5141086578369141, -0.4717264175415039, -0.41591984033584595, -0.3465694189071655, -0.26374712586402893, -0.16772118210792542, -0.05895814299583435, 0.06187836453318596, 0.19393208622932434, 0.33616170287132263, 0.487351655960083, 0.6461256742477417, 0.8109634518623352, 0.9802197217941284, 1.1521457433700562, 1.3249127864837646, 1.4966367483139038, 1.6654049158096313, 1.829302191734314, 1.9864403009414673, 2.1349828243255615, 2.2731735706329346, 2.399362564086914, 2.5120301246643066, 2.6098077297210693, 2.6915030479431152, 2.756113052368164, 2.802841901779175, 2.831112861633301, 2.8405754566192627, 2.831112861633301, 2.802841901779175, 2.756113052368164, 2.6915030479431152, 2.6098077297210693, 2.5120301246643066, 2.399362564086914, 2.2731735706329346, 2.1349828243255615, 1.9864403009414673, 1.829302191734314, 1.6654049158096313, 1.4966367483139038, 1.3249127864837646, 1.1521457433700562, 0.9802197217941284, 0.8109634518623352, 0.6461256742477417, 0.487351655960083, 0.33616170287132263, 0.19393208622932434, 0.06187836453318596, -0.05895814299583435, -0.16772118210792542, -0.26374712586402893, -0.3465694189071655, -0.41591984033584595, -0.4717264175415039, -0.5141086578369141, -0.5433695316314697, -0.5599846243858337, -0.5645893216133118, -0.5579633712768555, -0.5410138964653015, -0.5147561430931091, -0.48029395937919617, -0.43879905343055725, -0.3914893567562103, -0.33960819244384766, -0.284402996301651, -0.2271050363779068, -0.16891008615493774, -0.11096034944057465, -0.05432789772748947, -9.274395350058849e-17, 0.051133569329977036, 0.0982910767197609, 0.14080704748630524, 0.17813824117183685, 0.20986703038215637, 0.23570281267166138, 0.2554805874824524, 0.26915788650512695, 0.27680930495262146, 0.2786192297935486, 0.27487292885780334, 0.26594626903533936, 0.25229406356811523, 0.23443755507469177, 0.21295134723186493, 0.1884496808052063, 0.16157282888889313, 0.1329735815525055, 0.1033039465546608, 0.07320275157690048, 0.04328383877873421, 0.014125424437224865, -0.013739367946982384, -0.039830684661865234, -0.06372867524623871, -0.08507874608039856, -0.10359520465135574, -0.11906352639198303, -0.13134092092514038, -0.14035560190677643, -0.14610455930233002, -0.14865010976791382, -0.1481153666973114, -0.14467853307724, -0.13856664299964905, -0.13004827499389648, -0.11942609399557114, -0.10702886432409286, -0.09320346266031265, -0.07830678671598434, -0.06269800662994385, -0.046731047332286835, -0.030747657641768456, -0.01507099810987711, -5.148713261814661e-17, 0.014195486903190613, 0.027278704568743706, 0.039049070328474045, 0.04934436455368996, 0.058041974902153015, 0.06505925953388214, 0.07035301625728607, 0.0739181637763977, 0.07578572630882263, 0.07602003961801529, 0.07471556216478348, 0.07199306041002274, 0.0679955929517746, 0.06288408488035202, 0.056832920759916306, 0.05002541095018387, 0.04264933243393898, 0.03489265590906143, 0.02693951316177845, 0.018966492265462875, 0.011139304377138615, 0.003609905019402504, -0.0034859066363424063, -0.010030388832092285, -0.015925239771604538, -0.021092459559440613, -0.025474701076745987, -0.029035158455371857, -0.031756967306137085, -0.0336422435939312, -0.03471070155501366, -0.034998029470443726, -0.03455398976802826, -0.03344028815627098, -0.03172841668128967, -0.029497329145669937, -0.02683117426931858, -0.023817049339413643, -0.020542873069643974, -0.01709536649286747, -0.01355825737118721, -0.010010662488639355, -0.006525722797960043, -0.0031694755889475346, -1.6097297025390217e-17, 0.002933190669864416, 0.005589496344327927, 0.007937340065836906, 0.009954052045941353, 0.011625551618635654, 0.012945862486958504, 0.013916469179093838, 0.014545565471053123, 0.0148472273722291, 0.014840507879853249, 0.01454854290932417, 0.01399762649089098, 0.013216350227594376, 0.012234771624207497, 0.011083660647273064, 0.009793835692107677, 0.008395591750741005, 0.006918228697031736, 0.005389681551605463, 0.003836256917566061, 0.002282458357512951, 0.0007508968701586127, -0.0007377190049737692, -0.0021645394153892994, -0.0035124800633639097, -0.004766103345900774, -0.0059114787727594376, -0.006936056073755026,
//}

const (
	SampleRate = 900e3
	//SampleRate     = 56250
	BeaconOffset   = 143e3
	WorkDecimation = 16
	inputIQ        = "/media/ELTN/HAM/baseband/QO-100/SDRSharp_20190311_011013Z_739764000Hz_IQ_900e3.cfile"
	//inputIQ = "/media/ELTN/HAM/baseband/QO-100/SDRSharp_20190311_011013Z_739764000Hz_IQ_56250-N.cfile"
)

//var fullFFT *ui.FFTUI
//var partialFFT *ui.FFTUI

func main() {
	//log.Info("Starting UI")
	//ui.InitNK()
	//
	//fullFFT = ui.MakeFFTUI("Full FFT", SampleRate, 50e6)
	//fullFFT.SetWidth(1000)
	//fullFFT.SetHeight(600)
	//fullFFT.SetFFT2WTFRatio(0.6)
	//ui.AddComponent(fullFFT)
	//
	//partialFFT = ui.MakeFFTUI("Beacon", SampleRate/WorkDecimation, 0)
	//partialFFT.SetWidth(600)
	//partialFFT.SetHeight(600)
	//partialFFT.SetX(1000)
	//partialFFT.SetFFT2WTFRatio(0.6)
	//partialFFT.SetOffset(-10)
	//partialFFT.SetScale(8)
	//partialFFT.SetVerticalGridSteps(4)
	//ui.AddComponent(partialFFT)

	outSampleRate := SampleRate / WorkDecimation
	translatorTaps := dsp.MakeLowPass(64, SampleRate, outSampleRate/2-15e3, 15e3)
	slog.Info("Translator Taps Length: %d", len(translatorTaps))
	translator = dsp.MakeFrequencyTranslator(WorkDecimation, -BeaconOffset, SampleRate, translatorTaps)
	agc = dsp.MakeAttackDecayAGC(0.01, 0.2, 1, 10, 65536)
	costas = dsp.MakeCostasLoop2(0.01)
	interp = dsp.MakeFloatInterpolator(WorkDecimation)

	src := NewCFileFrontend(inputIQ)
	src.SetSampleRate(SampleRate)
	src.SetSamplesAvailableCallback(func(data SampleCallbackData) {
		sampleFifo.Add(data.ComplexArray)
	})

	slog.Info("Output Sample Rate: %f", outSampleRate)
	//fpsTicker := time.NewTicker(time.Second / 60)

	dspRunning = true
	go DSP()
	go Server()
	src.Start()

	var sig chan os.Signal
	sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	running := true

	for running {
		select {
		//case <-fpsTicker.C:
		//	if !ui.Loop() {
		//		running = false
		//		fpsTicker.Stop()
		//	}
		case <-sig:
			log.Info("Received SIGINT")
			running = false
		}
	}

	dspRunning = false
}
