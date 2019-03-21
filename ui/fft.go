package ui

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/racerxdl/segdsp/dsp"
	"github.com/racerxdl/segdsp/dsp/fft"
	"github.com/racerxdl/segdsp/tools"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"
)

const (
	initImageWidth  = 1024
	initImageHeight = 512
	initHGridSteps  = 6
	initVGridSteps  = 8

	initAcc       = 4.5
	initFFTSize   = 4096
	initFFTOffset = -20
	initFFTScale  = 4
)

type FFTUI struct {
	fftOffset              float32
	fftScale               float32
	fftCacheA              []float32
	fftCacheB              []float32
	acc                    float32
	fftSize                int32
	fftImage               *image.RGBA
	fftImageGraphicContext *draw2dimg.GraphicContext
	waterFallBuffers       []*image.RGBA
	sampleRate             float64
	centerFrequency        float64
	samplesCache           []complex64
	windowCache            []float64

	isUpdated bool

	title string

	parametersLock   sync.Mutex
	sampleBufferLock sync.Mutex
	drawLock         sync.Mutex

	x      float32
	y      float32
	width  float32
	height float32

	frameImg nk.Image
	frameTex int32

	hGridSteps int
	vGridSteps int

	fft2wtfR  float32
	fftHeight float32
	wtfHeight float32
}

func MakeFFTUI(title string, sampleRate float32, centerFrequency float32) *FFTUI {
	f := &FFTUI{
		fftOffset:       initFFTOffset,
		fftScale:        initFFTScale,
		fftCacheA:       make([]float32, 0),
		fftCacheB:       make([]float32, 0),
		acc:             initAcc,
		fftSize:         initFFTSize,
		sampleRate:      float64(sampleRate),
		centerFrequency: float64(centerFrequency),
		samplesCache:    make([]complex64, 0),
		windowCache:     dsp.BlackmanHarris(initFFTSize, 61),

		title: title,

		isUpdated:  true,
		x:          0,
		y:          0,
		width:      initImageWidth,
		height:     initImageHeight,
		frameTex:   -1,
		fft2wtfR:   0.5,
		hGridSteps: initHGridSteps,
		vGridSteps: initVGridSteps,

		parametersLock:   sync.Mutex{},
		sampleBufferLock: sync.Mutex{},
		drawLock:         sync.Mutex{},
	}

	f.refreshContext()
	f.UpdateSamples(make([]complex64, initFFTSize)) // Draw Buffer

	return f
}

func (c *FFTUI) refreshContext() {
	img := image.NewRGBA(image.Rect(0, 0, int(c.width), int(c.height)))
	gc := draw2dimg.NewGraphicContext(img)
	gc.SetFontData(draw2d.FontData{
		Name:   "FreeMono",
		Family: draw2d.FontFamilyMono,
		Style:  draw2d.FontStyleNormal,
	})
	gc.SetLineWidth(2)
	gc.SetStrokeColor(color.RGBA{R: 255, A: 255})
	gc.SetFillColor(color.RGBA{R: 0, A: 255})

	c.fftImage = img
	c.fftImageGraphicContext = gc

	c.fftHeight = c.height * c.fft2wtfR
	c.wtfHeight = c.height * (1 - c.fft2wtfR)

	c.waterFallBuffers = make([]*image.RGBA, int(c.wtfHeight))
}

func (c *FFTUI) UpdateSamples(data []complex64) {
	c.sampleBufferLock.Lock()
	defer c.sampleBufferLock.Unlock()

	// region Copy samples to local buffer soon as possible since this is probably called for a goroutine
	if len(c.samplesCache) < len(data) {
		c.samplesCache = make([]complex64, len(data))
	}
	copy(c.samplesCache, data)
	c.samplesCache = c.samplesCache[:len(data)]
	// endregion
	// region Lock and cache parameters for computing
	c.parametersLock.Lock()
	fftSize := c.fftSize
	fftOffset := c.fftOffset
	fftScale := c.fftScale
	acc := c.acc
	if c.fftCacheA != nil {
		if len(c.fftCacheA) != len(c.fftCacheB) {
			c.fftCacheB = make([]float32, len(c.fftCacheA))
		}
		copy(c.fftCacheB, c.fftCacheA)
	}
	c.parametersLock.Unlock()
	// endregion

	// region Compute FFT and Adjust for image
	if len(c.samplesCache) > int(fftSize) {
		c.samplesCache = c.samplesCache[:fftSize]
	}
	if len(c.windowCache) != len(c.samplesCache) {
		c.windowCache = dsp.BlackmanHarris(len(c.samplesCache), 61)
	}

	// Apply window
	for j := 0; j < len(c.samplesCache); j++ {
		s := c.samplesCache[j]
		c.samplesCache[j] = complex(real(s)*float32(c.windowCache[j]), imag(s)*float32(c.windowCache[j]))
	}

	fftResult := fft.FFT(c.samplesCache)
	fftReal := make([]float32, len(fftResult))
	if c.fftCacheB == nil || len(c.fftCacheB) != len(fftReal) {
		c.fftCacheB = make([]float32, len(fftReal))
		for i := 0; i < len(fftReal); i++ {
			c.fftCacheB[i] = 0
		}
	}

	var lastV = float32(0)
	for i := 0; i < len(fftResult); i++ {
		// Convert FFT to Power in dB
		var v = tools.ComplexAbsSquared(fftResult[i]) * float32(1.0/c.sampleRate)
		fftReal[i] = float32(10 * math.Log10(float64(v)))
		fftReal[i] = (c.fftCacheB[i]*(acc-1) + fftReal[i]) / acc
		if tools.IsNaN(fftReal[i]) || tools.IsInf(fftReal[i], 0) || tools.IsInf(fftReal[i], 1) {
			fftReal[i] = 0
		}
		if i > 0 {
			fftReal[i] = lastV*0.4 + fftReal[i]*0.6
		}
		lastV = fftReal[i]
		c.fftCacheB[i] = fftReal[i]
	}
	// endregion
	// region Update FFT Main Cache
	c.parametersLock.Lock()
	if c.fftCacheA == nil || len(c.fftCacheA) != len(c.fftCacheB) {
		c.fftCacheA = make([]float32, len(c.fftCacheB))
	}
	copy(c.fftCacheA, c.fftCacheB)
	c.parametersLock.Unlock()
	// endregion
	// region Draw Image and Save
	c.drawLock.Lock()
	defer c.drawLock.Unlock()

	if int(c.width) != c.fftImage.Bounds().Dx() || int(c.height) != c.fftImage.Bounds().Dy() || len(c.waterFallBuffers) != int(c.wtfHeight) {
		c.refreshContext()
	}

	widthScale := float32(len(fftReal)) / float32(c.width)
	gc := c.fftImageGraphicContext
	img := c.fftImage

	gc.Clear()
	gc.SetFontSize(32)
	gc.FillStringAt(c.title, 10, 10)
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.NRGBA{A: 255}}, image.ZP, draw.Src)

	lastX := float32(0)
	lastY := float32(0)

	maxVisible := float64(fftOffset) - float64(0)/float64(fftScale)
	minVisible := float64(fftOffset) - float64(c.fftHeight)/float64(fftScale)
	deltaVisible := math.Abs(maxVisible - minVisible)
	scale := 256 / deltaVisible

	wtfBuff := image.NewRGBA(image.Rect(0, 0, int(c.width), 1))

	for i := 0; i < len(fftReal); i++ {
		iPos := (i + len(fftReal)/2) % len(fftReal)
		s := float32(fftReal[iPos])
		v := float32((fftOffset)-s) * float32(fftScale)
		x := float32(i) / widthScale
		if i != 0 {
			DrawLine(lastX, lastY, x, v, color.NRGBA{R: 0, G: 127, B: 127, A: 255}, img)
		}

		waterFallZ := (float64(s) - minVisible) * scale
		if waterFallZ > 255 {
			waterFallZ = 255
		} else if waterFallZ < 0 {
			waterFallZ = 0
		}

		for i := lastX; i <= x; i++ { // For FFT Width < imgWidth
			wtfBuff.Set(int(i), 0, waterFallLut[uint8(waterFallZ)])
		}

		lastX = x
		lastY = v
	}

	if c.waterFallBuffers[0] != nil {
		if c.waterFallBuffers[0].Rect.Dx() != wtfBuff.Rect.Dx() { // Clear, wrong dimension
			c.waterFallBuffers = make([]*image.RGBA, int(c.wtfHeight))
		}
	}

	// Shift the waterfall forward, and add current as first.
	c.waterFallBuffers = append([]*image.RGBA{wtfBuff}, c.waterFallBuffers[:int(c.wtfHeight)-1]...)

	for i := 0; i < len(c.waterFallBuffers); i++ {
		line := c.waterFallBuffers[i]
		if line != nil {
			draw.Draw(img, image.Rect(0, i+int(c.fftHeight), line.Rect.Dx(), i+1+int(c.fftHeight)), line, image.ZP, draw.Src)
		}
	}

	c.drawGrid(gc, img, int(fftOffset), int(fftScale), int(c.width))
	c.drawChannelOverlay(gc, img, int(c.width))
	c.isUpdated = true
	// endregion
}

func (c *FFTUI) drawGrid(gc *draw2dimg.GraphicContext, img *image.RGBA, fftOffset, fftScale, width int) {
	var lineColor = color.NRGBA{R: 255, G: 127, B: 127, A: 127}
	gc.Save()
	gc.SetFontSize(10)
	gc.SetFillColor(color.NRGBA{R: 127, G: 127, B: 127, A: 255})

	var startFreq = c.centerFrequency - (c.sampleRate / 2)
	var hzPerPixel = c.sampleRate / float64(width)

	// region Draw dB Scale Grid
	for i := 0; i < c.hGridSteps; i++ {
		var y = float64(i) * (float64(c.fftHeight) / float64(c.hGridSteps))
		var dB = int(float64(fftOffset) - float64(y)/float64(fftScale))
		DrawLine(0, float32(y), float32(width), float32(y), lineColor, img)
		gc.FillStringAt(fmt.Sprintf("%d dB", dB), 5, y-5)
	}
	// endregion
	// region Draw Frequency Scale Grid
	for i := 0; i < c.vGridSteps; i++ {
		var x = math.Round(float64(i) * (float64(width) / float64(c.vGridSteps)))
		DrawLine(float32(x), 0, float32(x), float32(c.fftHeight), lineColor, img)
		var v = startFreq + float64(x)*hzPerPixel
		v2, unit := toNotationUnit(float32(v))
		gc.FillStringAt(fmt.Sprintf("%.2f %sHz", v2, unit), x+10, float64(c.fftHeight)-10)
	}
	// endregion
	gc.Restore()
}

func (c *FFTUI) drawChannelOverlay(gc *draw2dimg.GraphicContext, img *image.RGBA, width int) {
	//if demodulator != nil {
	//    var dp = demodulator.GetDemodParams()
	//    if dp != nil {
	//        var demodParams = dp.(demodcore.FMDemodParams)
	//        var channelWidth = demodParams.SignalBandwidth
	//
	//        var startX = FrequencyToPixelX(centerFreq-(channelWidth/2), float64(width))
	//        var endX = FrequencyToPixelX(centerFreq+channelWidth/2, float64(width))
	//
	//        for i := -1; i < 1; i++ {
	//            DrawLine(float32(startX)+float32(i), 0, float32(startX)+float32(i), fftHeight, color.NRGBA{192, 0, 0, 255}, img)
	//            DrawLine(float32(endX)+float32(i), 0, float32(endX)+float32(i), fftHeight, color.NRGBA{192, 0, 0, 255}, img)
	//        }
	//    }
	//}
}

func (c *FFTUI) Draw(win *glfw.Window, ctx *nk.Context) {
	nk.NkStyleSetFont(ctx, fonts["sans16"].Handle())
	bounds := nk.NkRect(c.x, c.y, c.width, c.height)
	update := nk.NkBegin(ctx, c.title, bounds, nk.WindowTitle)
	if update > 0 {
		if c.isUpdated {
			c.frameImg, c.frameTex = rgbaTex(c.frameTex, c.fftImage)
			c.isUpdated = false
		}
		nk.NkLayoutRowDynamic(ctx, c.height-60, 1)
		{
			nk.NkImage(ctx, c.frameImg)
		}
	}
	nk.NkEnd(ctx)
}

func (c *FFTUI) SetWidth(width float32) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	c.width = width
}

func (c *FFTUI) SetHeight(height float32) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	c.height = height
}

func (c *FFTUI) SetX(x float32) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	c.x = x
}

func (c *FFTUI) SetY(y float32) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	c.y = y
}

func (c *FFTUI) SetFFT2WTFRatio(r float32) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	if r > 1 || r < 0 {
		panic("ratio should be > 0 and < 1")
	}

	c.fft2wtfR = r
}

func (c *FFTUI) SetOffset(offset float32) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	c.fftOffset = offset
}

func (c *FFTUI) SetScale(scale float32) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()

	c.fftScale = scale
}

func (c *FFTUI) SetVerticalGridSteps(steps int) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	c.vGridSteps = steps
}

func (c *FFTUI) SetHorizontalGridSteps(steps int) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()
	c.hGridSteps = steps
}

func (c *FFTUI) SetCenterFrequency(centerFrequency float64) {
	c.parametersLock.Lock()
	defer c.parametersLock.Unlock()

	c.centerFrequency = centerFrequency
}

func (c *FFTUI) frequencyToPixelX(frequency, width float64) float64 {
	var hzPerPixel = c.sampleRate / float64(width)
	var delta = c.centerFrequency - float64(frequency)
	var centerX = float64(width / 2)

	return centerX + (delta / hzPerPixel)
}

func (c *FFTUI) Lock() {
	c.drawLock.Lock()
}

func (c *FFTUI) Unlock() {
	c.drawLock.Unlock()
}
