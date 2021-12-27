package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/inkyblackness/imgui-go/v4"
)

//17,28
const (
	WindowWidth  = 1920 / 1.5
	WindowHeight = 1080 / 1.5
	doJoystick   = false
	cameraType   = 0
)

// var JoystickPointer = nil

var dimensions = [2]float32{float32(WindowWidth), float32(WindowHeight)}

func main() {
	var FrameTime time.Duration
	//Make GL stuff happen on main thread
	runtime.LockOSThread()

	//Initialize GLFW
	err := glfw.Init()
	if err != nil {
		log.Fatalf("failed to initialize glfw: %v\n", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)

	//Create window
	window, err := glfw.CreateWindow(WindowWidth, WindowHeight, "GLSL Path tracing", nil, nil)
	if err != nil {
		glfw.Terminate()
		log.Fatalf("failed to create window: %v\n", err)
	}

	//Setup Window stuff
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	//Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version:", version)

	//Initialize Imgui
	context := imgui.CreateContext(nil)
	context.SetCurrent()
	defer context.Destroy()
	// Setup the Imgui renderer
	imguiRenderer, err := NewOpenGL3()
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer imguiRenderer.Dispose()

	imguiIO := imgui.CurrentIO()
	clipboard := CombinedClipboard{window}
	imguiIO.SetClipboard(clipboard)
	inputState := &ImguiInput{io: &imguiIO, time: 0}
	inputState.setKeyMapping()

	window.SetMouseButtonCallback(inputState.MouseButtonChange)
	window.SetScrollCallback(inputState.MouseScrollChange)
	window.SetKeyCallback(inputState.KeyChange)
	window.SetCharCallback(inputState.CharChange)

	//Setup tracer renderer
	InitRender()

	if doJoystick {
		log.Println("Using Joystick", glfw.Joystick1.GetName())

	}
	glfw.GetCurrentContext().SetKeyCallback(HandleKeys)
	//Loop
	for !window.ShouldClose() {
		timeStart := time.Now()
		frame++
		sameFrames++
		glfw.PollEvents()

		CreateNewIMGUIFrame(window, inputState)

		HandleClick(window)
		UpdateWorld()

		//Find id at mouse position
		//Draw imgui stuff
		imgui.Begin("Control")
		imgui.Text(fmt.Sprint("Frame:", frame))
		imgui.Text(fmt.Sprint("Frame Time:", FrameTime))
		imgui.Text(fmt.Sprintf("FPS %.2f", imgui.CurrentIO().Framerate()))

		imgui.Separator()
		imgui.Text(fmt.Sprint("Render Frames:", sameFrames))
		imgui.Text(fmt.Sprint("Samples Calculated:", samplesDone))
		imgui.Separator()
		imgui.Text("Camera Parameters")
		imgui.InputInt("Max Bounces", &maxBounces)
		imgui.InputInt("SamplesPerFrame", &SamplesPerFrame)
		if SamplesPerFrame > 100 {
			SamplesPerFrame = 100
		}

		changed := false
		changed = changed || imgui.DragFloat3V("origin", &origin, 0.001, -100, 100, "%g", 0)

		changed = imgui.DragFloatV("Focal Length", &focalLength, 0.001, -100, 100, "%g", 0) || changed
		changed = imgui.DragFloatV("Field of View", &fov, 0.05, 0, 180, "%g", 0) || changed

		changed = imgui.DragFloatV("Azimuth", &azimuth, 0.005, -2*3.14159, 2*3.14159+1, "%g", 0) || changed
		if azimuth < 0 {
			azimuth += 2 * 3.14169
		} else if azimuth > 2*3.14169 {
			azimuth -= 2 * 3.14169
		}
		changed = imgui.DragFloatV("Altitude", &altitude, 0.005, -3.14159/4, 3.14159/4, "%g", 0) || changed

		if changed {
			SceneChanged = true
		}
		if imgui.Button("NewFrame") {
			SceneChanged = true
		}
		imgui.Text(fmt.Sprint(resultFBHandle, idTexHandle, environmentHandle))
		imgui.Image(imgui.TextureID(idTexHandle), imgui.Vec2{200, 200})
		imgui.Image(imgui.TextureID(resultTexHandle), imgui.Vec2{200, 200})
		imgui.Image(imgui.TextureID(environmentHandle), imgui.Vec2{200, 200})
		imgui.End()
		//Safegaurds
		if maxBounces <= 0 {
			maxBounces = 1
		}
		if SamplesPerFrame <= 0 {
			SamplesPerFrame = 1
		}

		BuildObjectController()

		//Finish imgui
		imgui.Render()

		Draw()

		imguiRenderer.Render(dimensions, dimensions, imgui.RenderedDrawData())

		window.SwapBuffers()
		FrameTime = time.Since(timeStart)

	}
}

func GetMouseButtons123(window *glfw.Window) [3]bool {
	return [3]bool{window.GetMouseButton(glfw.MouseButton1) == glfw.Press,
		window.GetMouseButton(glfw.MouseButton2) == glfw.Press,
		window.GetMouseButton(glfw.MouseButton3) == glfw.Press}
}
func CreateNewIMGUIFrame(window *glfw.Window, inputState *ImguiInput) {
	cursorX, cursorY := window.GetCursorPos()
	mouseState := ImguiMouseState{
		MousePosX:  float32(cursorX),
		MousePosY:  float32(cursorY),
		MousePress: GetMouseButtons123(window)}

	inputState.NewFrame(WindowWidth, WindowHeight, glfw.GetTime(), window.GetAttrib(glfw.Focused) != 0, mouseState)
}

//Clipboard
type CombinedClipboard struct {
	window *glfw.Window
}

func (c CombinedClipboard) Text() (string, error) {
	return c.window.GetClipboardString(), nil
}
func (c CombinedClipboard) SetText(value string) {
	c.window.SetClipboardString(value)
}
