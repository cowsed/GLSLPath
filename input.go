package main

import (
	"fmt"
	"log"
	"math"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/inkyblackness/imgui-go/v4"
)

type KeyState int

const (
	KeyUp KeyState = iota
	KeyDown
)

var keysPressed map[glfw.Key]KeyState = map[glfw.Key]KeyState{}

func HandleInput() {

}
func HandleKeys(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		keysPressed[key] = KeyDown
	} else if action == glfw.Release {
		delete(keysPressed, key)
	}
}

func UpdateWorld() {
	var speed float32 = .04
	velocity = [3]float32{0, 0, 0}

	if cameraType == 0 {
		if keysPressed[glfw.KeyW] == KeyDown {
			velocity[2] += 1
		}
		if keysPressed[glfw.KeyS] == KeyDown {
			velocity[2] -= 1
		}

		if keysPressed[glfw.KeyA] == KeyDown {
			velocity[0] += 1
		}
		if keysPressed[glfw.KeyD] == KeyDown {
			velocity[0] -= 1
		}

		if keysPressed[glfw.KeyZ] == KeyDown {
			velocity[1] += 1
		}
		if keysPressed[glfw.KeyX] == KeyDown {
			velocity[1] -= 1
		}
		if keysPressed[glfw.KeyQ] == KeyDown {
			azimuth += 0.03
			SceneChanged = true

		}
		if keysPressed[glfw.KeyE] == KeyDown {
			azimuth -= 0.03

			SceneChanged = true

		}

		velocity = Normalize(velocity)
		velocity = Mul(velocity, speed)
		if doJoystick {
			jaxes := glfw.Joystick1.GetAxes()
			jspeed := .2 * jaxes[4]

			if len(jaxes) != 0 {
				if jaxes[1] != 0 {
					velocity[2] = jaxes[1] * jspeed
				}
				if jaxes[2] != 0 {
					velocity[1] = jaxes[2] * jspeed

				}
				if jaxes[3] != 0 {
					velocity[0] = -jaxes[3] * jspeed
				}
				if jaxes[0] != 0 {
					azimuth -= jaxes[0] * 0.05
					SceneChanged = true
				}

				if abs(altitude-(jaxes[5])) > 0.0001 {
					altitude = (jaxes[5])
					SceneChanged = true
				}

			}
		}
		velocity = rotateXV(velocity, -altitude)
		velocity = rotateYV(velocity, -azimuth)

	} else if cameraType == 1 {

	}
	if velocity[0] != 0 || velocity[1] != 0 || velocity[2] != 0 {
		SceneChanged = true
	}
	origin = Add(origin, velocity)
}

func abs(n float32) float32 {
	if n > 0 {
		return n
	}
	return -n
}
func rotateYV(v [3]float32, ang float32) [3]float32 {
	c := float32(math.Cos(float64(ang)))
	s := float32(math.Sin(float64(ang)))

	return [3]float32{
		v[0]*c - v[2]*s,
		v[1],
		v[2]*c + v[0]*s,
	}
}

func rotateXV(v [3]float32, ang float32) [3]float32 {
	c := float32(math.Cos(float64(ang)))
	s := float32(math.Sin(float64(ang)))

	return [3]float32{
		v[0],
		v[1]*c - v[2]*s,
		v[2]*c + v[1]*s,
	}
}

func Add(a, b [3]float32) [3]float32 {
	return [3]float32{
		a[0] + b[0],
		a[1] + b[1],
		a[2] + b[2],
	}
}
func Mul(a [3]float32, n float32) [3]float32 {
	return [3]float32{
		a[0] * n, a[1] * n, a[2] * n,
	}
}

func Normalize(a [3]float32) [3]float32 {
	l := float32(math.Sqrt(float64(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])))
	if l == 0 {
		return [3]float32{0, 0, 0}
	}
	return [3]float32{
		a[0] / l,
		a[1] / l,
		a[2] / l,
	}
}

func HandleClick(window *glfw.Window) {
	useMouse := !imgui.CurrentIO().WantCaptureMouse()
	if !useMouse {
		return
	}

	buttons := GetMouseButtons123(window)
	if buttons[0] {
		pixel := []float32{1, 0, 0, 0}
		cursorX, cursorY := window.GetCursorPos()
		log.Println("Picking at", cursorX, cursorY)
		gl.BindFramebuffer(gl.FRAMEBUFFER, idFBHandle)
		gl.ReadPixels(int32(cursorX), int32(WindowHeight-cursorY), 1, 1, gl.RGBA, gl.FLOAT, gl.Ptr(&pixel[0]))
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		fmt.Println(pixel)

		if pixel[0] == 1 {
			//Is the sky
			return
		}
		sphereID := int(pixel[2]*float32(NumSpheres) + .1)
		if sphereID < NumSpheres {
			currentObject = int32(sphereID)
		}
	}
}
