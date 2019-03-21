package ui

import (
	"fmt"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
	"github.com/quan-to/slog"
	"runtime"
)

const (
	winWidth  = 1600
	winHeight = 600

	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

var uiLog = slog.Scope("UI")
var fonts = make(map[string]*nk.Font)
var monoAtlas *nk.FontAtlas
var sansAtlas *nk.FontAtlas

var components = make([]DrawableComponent, 0)

var win *glfw.Window
var ctx *nk.Context

var state = &UIState{
	bgColor: nk.NkRgba(28, 48, 62, 255),
}

func init() {
	fontBytes := MustAsset("assets/FreeSans.ttf")
	loadedFont, err := truetype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}
	draw2d.RegisterFont(draw2d.FontData{
		Name:   "FreeMono",
		Family: draw2d.FontFamilyMono,
		Style:  draw2d.FontStyleNormal,
	}, loadedFont)

}

func InitNK() {
	runtime.LockOSThread()
	var err error
	if err := glfw.Init(); err != nil {
		uiLog.Fatal(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	win, err = glfw.CreateWindow(winWidth, winHeight, "QO100-DeDrift", nil, nil)
	if err != nil {
		uiLog.Fatal(err)
	}
	win.MakeContextCurrent()

	width, height := win.GetSize()

	if err := gl.Init(); err != nil {
		uiLog.Fatal("opengl initialization failed: %s", err)
	}
	gl.Viewport(0, 0, int32(width), int32(height))

	ctx = nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	// region Initialize Fonts
	var freeSansBytes = MustAsset("assets/FreeSans.ttf")
	var freeMonoBytes = MustAsset("assets/FreeMono.ttf")
	sansAtlas = nk.NewFontAtlas()
	nk.NkFontStashBegin(&sansAtlas)
	for i := 8; i <= 64; i += 2 {
		var fontName = fmt.Sprintf("sans%d", i)
		fonts[fontName] = nk.NkFontAtlasAddFromBytes(sansAtlas, freeSansBytes, float32(i), nil)
	}
	nk.NkFontStashEnd()
	monoAtlas = nk.NewFontAtlas()
	nk.NkFontStashBegin(&monoAtlas)
	for i := 8; i <= 64; i += 2 {
		var fontName = fmt.Sprintf("mono%d", i)
		fonts[fontName] = nk.NkFontAtlasAddFromBytes(monoAtlas, freeMonoBytes, float32(i), nil)
	}
	nk.NkFontStashEnd()
	// endregion

	nk.NkStyleSetFont(ctx, fonts["sans16"].Handle())
}

func gfxMain() {
	width, height := win.GetSize()
	nk.NkPlatformNewFrame()

	for _, v := range components {
		v.Lock()
		v.Draw(win, ctx)
		v.Unlock()
	}

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func Loop() bool {
	if win.ShouldClose() {
		nk.NkPlatformShutdown()
		glfw.Terminate()
		return false
	}
	glfw.PollEvents()
	gfxMain()

	return true
}

func AddComponent(component DrawableComponent) {
	components = append(components, component)
}
