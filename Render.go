package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	_ "embed"

	"github.com/go-gl/gl/v4.6-core/gl"
	_ "github.com/mdouchement/hdr/codec/rgbe"
)

var points []float32 = []float32{
	-1, 1, 0,
	-1, -1, 0,
	1, -1, 0,

	-1, 1, 0,
	1, 1, 0,
	1, -1, 0,
}

//go:embed Shaders/uniforms.glsl
var uniformSource string

//go:embed Shaders/util.glsl
var utilSource string

//go:embed Shaders/material.glsl
var materialSource string

//go:embed Shaders/main.frag
var fragmentMainSource string

var fragmentSource = uniformSource + utilSource + materialSource + fragmentMainSource

//go:embed Shaders/example.vert
var vertexSource string

var (
	vboHandle         uint32
	vaoHandle         uint32
	programHandle     uint32
	resultFBHandle    uint32
	resultTexHandle   uint32
	idFBHandle        uint32
	idTexHandle       uint32
	environmentHandle uint32
)

func Draw() {
	if SceneChanged {
		sameFrames = 0
		samplesDone = 0
		SceneChanged = false
	}
	samplesDone += uint32(SamplesPerFrame)
	UpdateUniforms()

	loc := gl.GetUniformLocation(programHandle, gl.Str("render_stage\x00"))

	//Draw ray tracing
	gl.BindFramebuffer(gl.FRAMEBUFFER, resultFBHandle)

	gl.Uniform1i(loc, 2)
	gl.ClearColor(1, 0, 0, 1)

	if SceneChanged {

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(1, 0, 1, 1)

	gl.Viewport(0, 0, WindowWidth, WindowHeight)

	gl.UseProgram(programHandle)
	gl.BindVertexArray(vaoHandle)

	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(points)/3))

	////Draw ids
	gl.BindFramebuffer(gl.FRAMEBUFFER, idFBHandle)

	gl.Uniform1i(loc, 1)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0, 1, 1, 1)

	gl.Viewport(0, 0, WindowWidth, WindowHeight)

	gl.UseProgram(programHandle)
	gl.BindVertexArray(vaoHandle)

	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(points)/3))

	//Draw results to screen
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	gl.Uniform1i(loc, 0)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0, 1, 0, 1)

	gl.Viewport(0, 0, WindowWidth, WindowHeight)

	gl.UseProgram(programHandle)
	gl.BindVertexArray(vaoHandle)

	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(points)/3))

}

func InitRender() {
	//Save what the final shader was for debugging
	f, err := os.Create("Shaders/output.glsl")
	if err != nil {
		log.Fatal(err)
	}
	f.Write([]byte(fragmentSource))
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	gl.GenBuffers(1, &vboHandle)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboHandle)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	gl.GenVertexArrays(1, &vaoHandle)
	gl.BindVertexArray(vaoHandle)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboHandle)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	//Create Result FB
	gl.GenFramebuffers(1, &resultFBHandle)

	//Create result texture
	gl.GenTextures(1, &resultTexHandle)
	gl.BindTexture(gl.TEXTURE_2D, resultTexHandle)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, WindowWidth, WindowHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	//Connect result FB and texture
	gl.BindFramebuffer(gl.FRAMEBUFFER, resultFBHandle)
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, resultTexHandle, 0)

	//Generate id FB
	gl.GenFramebuffers(1, &idFBHandle)

	//Create id texture
	gl.GenTextures(1, &idTexHandle)
	gl.BindTexture(gl.TEXTURE_2D, idTexHandle)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, WindowWidth, WindowHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	//Connect id FB and texture
	gl.BindFramebuffer(gl.FRAMEBUFFER, idFBHandle)
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, idTexHandle, 0)

	//Bring back the default
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindTexture(gl.TEXTURE_2D, gl.TEXTURE0)

	//Builds shader program
	BuildProgram()

	//Load HDRI
	//environmentHandle = LoadHDRI("Environments/cannon2.jpg")
}
func LoadHDRI(fname string) uint32 {
	log.Println("Loading Image")
	fi, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	m, _, err := image.Decode(fi)
	if err != nil {
		panic(err)
	}
	log.Printf("Loading image %dx%d\n", m.Bounds().Dx(), m.Bounds().Dy())
	log.Println(m.Bounds())
	imgwidth := m.Bounds().Dx()
	imgheight := m.Bounds().Dy()

	i := image.NewRGBA(m.Bounds())
	for x := 0; x < imgwidth; x++ {
		for y := 0; y < imgheight; y++ {
			i.Set(x, y, m.At(x, y))
		}
	}

	f, _ := os.Create("f.png")
	png.Encode(f, i)
	f.Close()

	var texHandle uint32
	gl.GenTextures(1, &texHandle)
	internalFmt := int32(gl.SRGB_ALPHA)
	format := uint32(gl.RGBA)
	width := int32(m.Bounds().Dx())
	height := int32(m.Bounds().Dy())
	pixType := uint32(gl.UNSIGNED_BYTE)
	dataPtr := gl.Ptr(i.Pix)

	gl.BindTexture(gl.TEXTURE_2D, texHandle)
	//gl.TexParameteri(target, gl.TEXTURE_WRAP_R, wrapR)
	//gl.TexParameteri(target, gl.TEXTURE_WRAP_S, wrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // minification filter
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR) // magnification filter

	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFmt, width, height, 0, format, pixType, dataPtr)

	gl.GenerateMipmap(texHandle)

	gl.BindTexture(gl.TEXTURE_2D, 0)
	log.Println("Finished Loading HDRI")

	return texHandle
}

//OpenGL Stuff

func BuildProgram() {
	log.Println("Building Shader Program")
	//Delete old program
	gl.DeleteProgram(programHandle)

	//Compile Vertex Shader
	vertexShader, err := compileShader(vertexSource+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	//Compile Fragment Shader
	fragmentShader, err := compileShader(fragmentSource+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		log.Fatal(err)
		return
	}

	//Create and Link Program
	programHandle = gl.CreateProgram()
	gl.AttachShader(programHandle, vertexShader)
	gl.AttachShader(programHandle, fragmentShader)
	gl.LinkProgram(programHandle)

	//Release programs
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	//Check Link Errors
	var isLinked int32
	gl.GetProgramiv(programHandle, gl.LINK_STATUS, &isLinked)
	if isLinked == gl.FALSE {
		var maxLength int32
		gl.GetProgramiv(fragmentShader, gl.INFO_LOG_LENGTH, &maxLength)

		infoLog := make([]uint8, maxLength+1) //[bufSize]uint8{}
		gl.GetShaderInfoLog(fragmentShader, maxLength, &maxLength, &infoLog[0])

		log.Println("Link Result\n", string(infoLog), "")
		return
	}
	log.Println("Shaders built correctly")
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		shaderString := "Unknown"
		if shaderType == gl.FRAGMENT_SHADER {
			shaderString = "Fragment"
		} else if shaderType == gl.VERTEX_SHADER {
			shaderString = "Vertex"
		}
		return 0, fmt.Errorf("failed to compile type %s:\nLog:\n%v", shaderString, log[:len(log)-2])
	}

	return shader, nil
}
