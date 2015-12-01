package main

import (
	"fmt"
	"github.com/akiross/gogloo"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
	"github.com/veandco/go-sdl2/sdl_ttf"
	"runtime"
	"time"
	"unsafe"
)

// Shaders
var vertShader, fragShader *gogloo.ShaderObject
var canvasVao *gogloo.VertexArrayObject
var shadProg *gogloo.ProgramObject

func MakeFragmentShader(tree string) string {
	shader := `#version 130
in vec2 fragUV;
out vec3 outColor;

#define FILL(col) outColor = col;

#define HSPLIT(above, below) \
	if (lim.z < fragUV.y && fragUV.y <= (lim.z + lim.w) * 0.5) {\
		lim = vec4(lim.xy, lim.z, (lim.z + lim.w) * 0.5);\
		below\
	}\
	else {\
		lim = vec4(lim.xy, (lim.z + lim.w) * 0.5, lim.w);\
		above\
	}
#define VSPLIT(left, right) \
	if (lim.x < fragUV.x && fragUV.x <= (lim.x + lim.y) * 0.5) {\
		lim = vec4(lim.x, (lim.x + lim.y) * 0.5, lim.zw);\
		left\
	}\
	else {\
		lim = vec4((lim.x + lim.y) * 0.5, lim.x, lim.zw);\
		right\
	}
void main() {
	vec4 lim = vec4(0, 1, 0, 1);`

	return shader + tree + "\n}"
}

func checkGLError(where string) {
	switch gl.GetError() {
	case gl.NO_ERROR:
		// All fine! :)
	case gl.INVALID_ENUM:
		fmt.Println("ERROR: GL_INVALID_ENUM", where)
	case gl.INVALID_VALUE:
		fmt.Println("ERROR: GL_INVALID_VALUE", where)
	case gl.INVALID_OPERATION:
		fmt.Println("ERROR: GL_INVALID_OPERATION", where)
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		fmt.Println("ERROR: GL_INVALID_ FB OP", where)
	case gl.OUT_OF_MEMORY:
		fmt.Println("ERROR: out of memory", where)
	case gl.STACK_UNDERFLOW:
		fmt.Println("ERROR: stack underflow", where)
	case gl.STACK_OVERFLOW:
		fmt.Println("ERROR: stack overflow", where)
	default:
		fmt.Println("ERROR: Another unknown error!", where)
	}
}

func saveScreen(fname string, x, y, w, h int) {
	var data []byte = make([]byte, w*h*3)

	if gl.GetError() != gl.NO_ERROR {
		fmt.Println("BAD STATE BEFORE SAVING! :(")
	}

	gl.ReadPixels(int32(x), int32(y), int32(w), int32(h), gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(data))

	if gl.GetError() != gl.NO_ERROR {
		fmt.Println("BAD STATE after SAVING! :(")
	}

	// Flip data vertically
	if true {
		s := 3 * w // Stride
		for r := 0; r < h/2; r++ {
			nr := h - r - 1
			for i := 0; i < s; i++ {
				data[r*s : r*s+s][i], data[nr*s : nr*s+s][i] = data[nr*s : nr*s+s][i], data[r*s : r*s+s][i]
			}
		}
	}

	surf, err := sdl.CreateRGBSurfaceFrom(unsafe.Pointer(&data[0]), w, h, 24, w*3, 0x000000FF, 0x0000FF00, 0x00FF0000, 0)
	if err != nil {
		panic(err)
	}

	surf.SaveBMP(fname)
	fmt.Println("ZAVED!")
}

func main() {
	// Random crashes, this was suggested to be used
	runtime.LockOSThread()

	//progs = make([]*ProgramObject, 2)

	var wTitle string = "Recursive Image Representation"
	var wWidth, wHeight int = 500, 500
	var running bool

	// Initialize SDL
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	fmt.Println("SDL initialized")

	// Initialize SDL_image
	if loaded := img.Init(img.INIT_PNG); loaded&img.INIT_PNG == 0 {
		panic(fmt.Sprint("Cannot initialize PNG images:", img.GetError()))
	}
	defer img.Quit()
	fmt.Println("Image initialized")

	// Initialize SDL_ttf
	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()
	fmt.Println("TTF initialized")

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		panic(err)
	}
	fmt.Println("OpenGL initialized")

	// Create the OpenGL-capable window
	win, err := sdl.CreateWindow(
		wTitle,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		wWidth, wHeight,
		sdl.WINDOW_OPENGL)

	if err != nil {
		panic(err)
	}
	defer win.Destroy()
	fmt.Println("Window created")

	// Setup OpenGL context
	var ctx sdl.GLContext
	ctx, err = sdl.GL_CreateContext(win)
	if err != nil {
		panic(err)
	}
	defer sdl.GL_DeleteContext(ctx)
	fmt.Println("OpenGL context created")

	// Settings
	initGL(wWidth, wHeight)

	// Load a font
	font, err := ttf.OpenFont("/home/akiross/Resources/python-tutorials/assets/fonts/FreeSans.ttf", 24)
	if err != nil {
		panic(err)
	}
	defer font.Close()
	fmt.Println("Font: FreeSans loaded")

	/* Example text rendering
	// Render font to a surface
	sfondo, _ := sdl.CreateRGBSurface(0, 300, 300, 32, 0, 0, 0, 0)
	color := sdl.Color{255, 255, 255, 255}
	textSurf, err := font.RenderUTF8_Blended("Hèllo W©rld!", color)
	if err != nil {
		panic(err)
	}
	fmt.Println("Example text rendered")
	textSurf.Blit(nil, sfondo, nil)
	sfondo.SaveBMP("example_surface.bmp")
	*/

	// Set a deadline for the reading, so that it's non-blocking
	fmt.Println("Initialization sequence terminated!\nStarting main loop :)")

	framePerSecond := int64(1)
	timePerFrame := time.Duration(int64(time.Second) / framePerSecond)

	running = true
	for running {
		// Time at beginning of frame
		timeBegin := time.Now()

		// Process all pending events
		//mouseMoved = false
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		draw()
		// Show on window
		sdl.GL_SwapWindow(win)

		//gl.Finish()
		//gl.ReadBuffer(gl.FRONT)
		// Save screen to image (after swap, we capture GL_FRONT)
		saveScreen("ciccia.bmp", 0, 0, 500, 500)

		//os.Exit(0)

		//saveImage("prova.ppm", data)

		// Time after frame actions and rendering
		timeDelta := time.Since(timeBegin)
		// If frame took less than required, wait some time
		if timeDelta < timePerFrame {
			wt := (timePerFrame - timeDelta) / time.Millisecond
			sdl.Delay(uint32(wt))
		} else {
			fmt.Println("Frame took too long:", timeDelta)
		}
	}
}

var offscreenFBO *gogloo.FrameBufferObject
var offscreenRBO *gogloo.RenderBufferObject
var offscreenTex *gogloo.TextureObject

func initGL(winW, winH int) {
	gl.Enable(gl.TEXTURE_2D)

	gl.ClearColor(0.2, 0.2, 0.3, 1.0)

	gl.Viewport(0, 0, int32(winW), int32(winH))
	offscreenFBO = gogloo.CreateFBO()
	offscreenFBO.Bind()
	// Render to texture or to RBO?
	if false {
		offscreenTex = gogloo.CreateEmptyTO(100, 100)
		offscreenFBO.SetTexture(offscreenTex)
	} else {
		offscreenRBO = gogloo.CreateRBO()
		offscreenRBO.Bind()
		offscreenRBO.Storage(gl.RGB, 100, 100)
		offscreenFBO.SetRenderBuffer(offscreenRBO)
	}

	if !offscreenFBO.StatusComplete() {
		fmt.Println("INCOMPLETE FBO!!!")
		offscreenFBO.StatusError()
	}

	offscreenFBO.Unbind()
	// Shader sources
	vertShaderSrc := `
#version 130

in vec3 vertPos;
in vec2 vertUV;

out vec2 fragUV;

void main() {
	gl_Position = vec4(vertPos, 1.0f);
	fragUV = vertUV;
}
`
	fragShaderSrc := MakeFragmentShader("HSPLIT(VSPLIT(FILL(vec3(0.3)), FILL(vec3(1.0))), HSPLIT(VSPLIT(FILL(vec3(1, 0, 0)), FILL(vec3(0, 0, 1))), FILL(vec3(1, 1, 0))))")
	glslCompStartTime := time.Now()

	vertShader = gogloo.CreateShader("VShad", vertShaderSrc, gl.VERTEX_SHADER)
	if vertShader.Compile() != nil {
		panic(vertShader.GetLog())
	}
	fragShader = gogloo.CreateShader("FShad", fragShaderSrc, gl.FRAGMENT_SHADER)
	if fragShader.Compile() != nil {
		panic(fragShader.GetLog())
	}

	shadProg = gogloo.CreateProgramObject()
	shadProg.Attach(vertShader)
	shadProg.Attach(fragShader)

	vaaPos := shadProg.BindAttribLocation(0, "vertPos")
	vaaUV := shadProg.BindAttribLocation(1, "vertUV")

	if shadProg.Link() != nil {
		panic(shadProg.GetLog())
	}

	compileDuration := time.Since(glslCompStartTime)

	fmt.Println("To compile the shader it took", compileDuration)

	/*
		progs[0], vaas = PrepareProgram(
			[]string{shaders[0]}, // Static shader
			[]string{shaders[2]},
			[]string{"vertPos", "vertUV", "vertCol"},
			[]uint32{0, 1, 2})
	*/

	shadProg.Use()

	const N float32 = -0.5
	const P float32 = 0.5

	// Setup the "canvas": a simple quad
	canvasData := []float32{
		N, N, 0, 0.0, 0.0,
		P, N, 0, 1.0, 0.0,
		N, P, 0, 0.0, 1.0,
		P, P, 0, 1.0, 1.0,
	}

	canvasVao = gogloo.CreateVAO(canvasData)

	// Bind VAO
	canvasVao.Bind()

	// Bind VBO
	canvasVao.Data.Bind()

	vaaPos.Enable()
	vaaUV.Enable()

	vaaPos.SetPointer(3, gl.FLOAT, false, 5*4, 0)
	vaaUV.SetPointer(2, gl.FLOAT, false, 5*4, 3*4)
}

func draw() {
	if true {
		offscreenFBO.Bind()
		gl.Viewport(0, 0, 100, 100)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		gl.Finish()
		saveScreen("fbo100.bmp", 0, 0, 100, 100)
		offscreenFBO.Unbind()
		gl.Viewport(0, 0, 500, 500)
	}

	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	/*
		if false {
			gl.Finish()

			pixData := offscreenFBO.ReadPixelsRGB(0, 0, 100, 100)
			pixData = pixData

		}
	*/
}
