package main

import (
	"math"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/inkyblackness/imgui-go/v4"
)

const NumSpheres = 12
const NumMaterials = 10

var SceneChanged bool
var samplesDone uint32 = 0

//Camera
var (
	azimuth  float32 = 3.14159
	altitude float32 = 0.0
)

var (
	sameFrames            = 0
	frame                 = 0
	maxBounces      int32 = 20
	SamplesPerFrame int32 = 3

	//viewIds bool = false

	focalLength float32 = 1.0
	fov         float32 = 90
	origin              = [3]float32{0, 0, 0}
	lookat              = [3]float32{0, 0, -1}
	camup               = [3]float32{0, 1, 0}
	velocity            = [3]float32{0, 0, 0}

	spherePositions = [NumSpheres][3]float32{{0, -1000.5, -1}, {0, 0, -1}, {-1, 0, -1}, {1, 0, -1}, {-.6, -.4, -1}, {.6, -.4, -.8}, {-.6, -.4, -.8}, {-.7, -.4, -.8}, {.7, -.4, -.8}}
	sphereRadii     = [NumSpheres]float32{1000, .5, .5, .5, .1, .1, .1, .1, .1}
	sphereMaterials = [NumSpheres]int32{0, 1, 2, 3, 3, 4, 5, 1, 1, 1}

	materialColors = [NumMaterials][3]float32{{.8, .8, 0}, {.7, .3, .3}, {.8, .8, .8}, {.8, .6, .2}, {0, 1, 0}, {0, 0, 1}}

	materialFuzz  = [NumSpheres]float32{0, 0, 0, 0, .1, .1, .1, .1}
	materialTypes = [NumSpheres]int32{0, 0, 1, 1, 1, 1, 0, 0, 0}
)

func dirFromAngles(azimuth, altitude float32) [3]float32 {
	dir := [3]float32{
		float32(math.Sin(float64(azimuth)) * math.Cos(float64(altitude))),
		float32(math.Sin(float64(altitude))),
		float32(math.Cos(float64(azimuth)) * math.Cos(float64(altitude))),
	}

	return dir
}

func UpdateUniforms() {

	var loc int32
	//Render Parameters
	loc = gl.GetUniformLocation(programHandle, gl.Str("frame\x00"))
	gl.Uniform1f(loc, float32(frame))

	loc = gl.GetUniformLocation(programHandle, gl.Str("sameFrame\x00"))
	gl.Uniform1f(loc, float32(sameFrames))

	//Send texture
	gl.BindTexture(gl.TEXTURE_2D, resultTexHandle)

	loc = gl.GetUniformLocation(programHandle, gl.Str("previousResult\x00"))
	gl.Uniform1ui(loc, resultTexHandle)
	gl.BindTexture(gl.TEXTURE_2D, gl.TEXTURE0)

	loc = gl.GetUniformLocation(programHandle, gl.Str("maxBounces\x00"))
	gl.Uniform1i(loc, maxBounces)

	loc = gl.GetUniformLocation(programHandle, gl.Str("SamplesPerFrame\x00"))
	gl.Uniform1i(loc, SamplesPerFrame)

	loc = gl.GetUniformLocation(programHandle, gl.Str("windowDimensions\x00"))
	gl.Uniform2f(loc, dimensions[0], dimensions[1])

	//Camera Parameters
	loc = gl.GetUniformLocation(programHandle, gl.Str("origin\x00"))
	gl.Uniform3f(loc, origin[0], origin[1], origin[2])

	loc = gl.GetUniformLocation(programHandle, gl.Str("camup\x00"))
	gl.Uniform3f(loc, camup[0], camup[1], camup[2])

	lookat := dirFromAngles(azimuth, altitude)
	lookat = Add(lookat, origin)
	loc = gl.GetUniformLocation(programHandle, gl.Str("lookat\x00"))
	gl.Uniform3f(loc, lookat[0], lookat[1], lookat[2])

	loc = gl.GetUniformLocation(programHandle, gl.Str("focal_length\x00"))
	gl.Uniform1f(loc, focalLength)

	loc = gl.GetUniformLocation(programHandle, gl.Str("field_of_view\x00"))
	gl.Uniform1f(loc, fov)

	//Sphere parameters
	loc = gl.GetUniformLocation(programHandle, gl.Str("spherePositions\x00"))
	gl.Uniform3fv(loc, int32(len(spherePositions)), &spherePositions[0][0])

	loc = gl.GetUniformLocation(programHandle, gl.Str("sphereRadii\x00"))
	gl.Uniform1fv(loc, int32(len(sphereRadii)), &sphereRadii[0])

	loc = gl.GetUniformLocation(programHandle, gl.Str("sphereMaterials\x00"))
	gl.Uniform1iv(loc, int32(len(sphereMaterials)), &sphereMaterials[0])

	//Material Parameters
	loc = gl.GetUniformLocation(programHandle, gl.Str("materialColors\x00"))
	gl.Uniform3fv(loc, int32(len(materialColors)), &materialColors[0][0])
	//materialFuzziness
	loc = gl.GetUniformLocation(programHandle, gl.Str("materialFuzziness\x00"))
	gl.Uniform1fv(loc, int32(len(materialFuzz)), &materialFuzz[0])

	loc = gl.GetUniformLocation(programHandle, gl.Str("materialTypes\x00"))
	gl.Uniform1iv(loc, int32(len(materialTypes)), &materialTypes[0])

}

var currentObject int32 = 0

func BuildObjectController() {
	imgui.Begin("Objects")
	imgui.InputInt("Current Object", &currentObject)
	if int(currentObject) >= len(spherePositions) {
		currentObject -= int32(len(spherePositions))
	} else if currentObject < 0 {
		currentObject = (currentObject * -1) % int32(len(spherePositions))
	}

	changed := false
	changed = imgui.DragFloat3V("Position", &spherePositions[currentObject], 0.001, -10000, 10000, "%g", 0) || changed
	changed = imgui.DragFloatV("Radius", &sphereRadii[currentObject], 0.001, 0, 10000, "%g", 0) || changed
	changed = imgui.InputInt("Material Index", &sphereMaterials[currentObject]) || changed
	sphereMaterials[currentObject] %= NumMaterials
	if sphereMaterials[currentObject] < 0 {
		sphereMaterials[currentObject] = 0
	}

	currentMaterial := sphereMaterials[currentObject]
	imgui.Separator()
	imgui.Text("Material")
	if int(currentMaterial) >= len(materialColors) {
		currentMaterial -= int32(len(materialColors))
	} else if currentMaterial < 0 {
		currentMaterial = (currentMaterial * -1) % int32(len(materialColors))
	}
	//imgui.ColorEdit3V("Albedo", &materialColors[currentMaterial], im`)
	changed = imgui.ColorPicker3("Albedo", &materialColors[currentMaterial]) || changed
	changed = imgui.DragFloatV("Fuzziness", &materialFuzz[currentMaterial], 0.001, 0, 5, "%g", 0) || changed
	changed = imgui.InputInt("Material Type", &materialTypes[currentMaterial]) || changed

	if changed {
		SceneChanged = true
	}
	imgui.End()
}
