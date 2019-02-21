package main

// typedef unsigned char Uint8;
// void SineWave(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"reflect"
	"unsafe"

	chip8 "github.com/abdullah2993/go-chip8"
	"github.com/veandco/go-sdl2/sdl"
)

const cW = 64
const cH = 32
const wW = 64 * 8
const wH = 32 * 8

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Chip8 Emulator", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		wW, wH, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	surface.FillRect(nil, 0)

	sur, err := sdl.CreateRGBSurfaceWithFormat(0, cW, cH, 8, sdl.PIXELFORMAT_RGB888)
	if err != nil {
		panic(err)
	}
	gc := &sdlGC{
		ts: sur,
		ws: surface,
		w:  window,
	}
	initAudio()
	c, err := chip8.NewChip8FromROM(gc, sdlC{}, sdlC{}, os.Args[1])
	// c, err := chip8.NewChip8(gc, sdlC{}, sdlC{}, []uint8{0xA0, 75, 0xD0, 0x05})
	if err != nil {
		panic(err)
	}

	s := c.Loop()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				s()
				running = false
				break
			}

		}
	}
}

type sdlGC struct {
	ts *sdl.Surface
	ws *sdl.Surface
	w  *sdl.Window
}

func (s *sdlGC) Clear() {
	s.ws.FillRect(nil, 0)
}

func (s *sdlGC) Draw(pd [cH][cW]uint8) {
	for i := 0; i < cH; i++ {
		for j := 0; j < cW; j++ {
			c := color.Black
			if pd[i][j] == 1 {
				c = color.White
			}
			s.ts.Set(j, i, c)
		}
	}
	s.ts.BlitScaled(nil, s.ws, nil)
	s.w.UpdateSurface()
}

type sdlC struct{}

func (sdlC) Sound(on bool) {
	sdl.PauseAudio(!on)
}
func (sdlC) IsPressed(k uint8) bool {
	fmt.Println("Wants key: ", km[k])
	ks := sdl.GetKeyboardState()
	for i := 0; i < len(ks); i++ {
		if ks[i] == 1 && sdl.GetKeyName(sdl.GetKeyFromScancode(sdl.Scancode(i))) == km[k] {
			return true
		}
	}
	return false
}
func (sdlC) WaitKey() uint8 {
	fmt.Println("Waiting key")
	for {
		ks := sdl.GetKeyboardState()
		for i := 0; i < len(km); i++ {
			sc := sdl.GetScancodeFromName(km[i])
			if ks[sc] == 1 {
				fmt.Println("got", i)
				return uint8(i)
			}
		}
	}
}

const (
	toneHz   = 440
	sampleHz = 48000
	dPhase   = 2 * math.Pi * toneHz / sampleHz
)

//export SineWave
func SineWave(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]C.Uint8)(unsafe.Pointer(&hdr))

	var phase float64
	for i := 0; i < n; i += 2 {
		phase += dPhase
		sample := C.Uint8((math.Sin(phase) + 0.999999) * 128)
		buf[i] = sample
		buf[i+1] = sample
	}
}

func initAudio() {
	spec := &sdl.AudioSpec{
		Freq:     sampleHz,
		Format:   sdl.AUDIO_U8,
		Channels: 2,
		Samples:  sampleHz,
		Callback: sdl.AudioCallback(C.SineWave),
	}
	if err := sdl.OpenAudio(spec, nil); err != nil {
		log.Println(err)
		return
	}
}

var km = [16]string{
	"1", "2", "3", "4",
	"Q", "W", "E", "R",
	"A", "S", "D", "F",
	"Z", "X", "C", "V",
}
