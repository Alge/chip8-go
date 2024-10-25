package main

import (
	"fmt"
	"github.com/Alge/chip8-go/chip8"
	sdl "github.com/veandco/go-sdl2/sdl"
	"log"
	"time"
  "os"
)

// Default Chip8 resolution
const CHIP_8_WIDTH int32 = 64
const CHIP_8_HEIGHT int32 = 32
const speed float64 = 700
const modifier int32 = 20

//const fileName string = "roms/chip8-picture.ch8"
//const fileName string = "roms/out.ch8"

func main() {
	fmt.Println("Starting up!")

	e := chip8.New()

	//chip8.LoadRom(e, "roms/chip8-picture.ch8")
	//chip8.LoadRom(e, "roms/pong.c8")
	//chip8.LoadRom(e, "roms/invaders.c8")
	//chip8.LoadRom(e, "roms/filter.c8")
	//chip8.LoadRom(e, "roms/out.ch8")

  fileName := os.Args[1]

	chip8.LoadRom(e, fileName)

	fmt.Printf("PC: 0x%X\n", e.PC)
	fmt.Print("RAM:")
  e.PrintRAM()
	fmt.Println(e.Registers)

	// Run untill we reach a "0" byte. Not technically to spec...
	fmt.Println("Starting execution")
	fmt.Println()

	// Initialize sdl2
	if sdlErr := sdl.Init(sdl.INIT_EVERYTHING); sdlErr != nil {
		panic(sdlErr)
	}
	defer sdl.Quit()

	// Create window, chip8 resolution with given modifier
	window, windowErr := sdl.CreateWindow("Chip 8 - "+fileName, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, CHIP_8_WIDTH*modifier, CHIP_8_HEIGHT*modifier, sdl.WINDOW_SHOWN)
	if windowErr != nil {
		panic(windowErr)
	}
	defer window.Destroy()

	// Create render surface
	canvas, canvasErr := sdl.CreateRenderer(window, -1, 0)
	if canvasErr != nil {
		panic(canvasErr)
	}
	defer canvas.Destroy()

	num_empty := 0

	for {
		if false && e.RAM[e.PC] == 0 {
			num_empty++
			if num_empty > 10 {
				break
			}
		} else {
			num_empty = 0
		}
		e.Tick()
    //e.PrintRAM()

		if e.Draw {
			log.Println("Re-drawing screen")
			e.Draw = false
			canvas.SetDrawColor(0, 0, 0, 255)
			canvas.Clear()

			for y, line := range e.Display {
				for x, pixel := range line {
					if pixel == 0 {
						canvas.SetDrawColor(0, 0, 0, 255)
					} else {
						canvas.SetDrawColor(255, 255, 255, 255)
					}

					canvas.FillRect(&sdl.Rect{
						Y: int32(y) * modifier,
						X: int32(x) * modifier,
						W: modifier,
						H: modifier,
					})

				}
				//log.Println(line)
			}
			canvas.Present()
		}

		// Poll for Quit and Keyboard events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch et := event.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.KeyboardEvent:
				if et.Type == sdl.KEYUP {
					switch et.Keysym.Sym {
					case sdl.K_1:
						e.Keyboard[0x1] = false
					case sdl.K_2:
						e.Keyboard[0x2] = false
					case sdl.K_3:
						e.Keyboard[0x3] = false
					case sdl.K_4:
						e.Keyboard[0xC] = false
					case sdl.K_q:
						e.Keyboard[0x4] = false
					case sdl.K_w:
						e.Keyboard[0x5] = false
					case sdl.K_e:
						e.Keyboard[0x6] = false
					case sdl.K_r:
						e.Keyboard[0xD] = false
					case sdl.K_a:
						e.Keyboard[0x7] = false
					case sdl.K_s:
						e.Keyboard[0x8] = false
					case sdl.K_d:
						e.Keyboard[0x9] = false
					case sdl.K_f:
						e.Keyboard[0xE] = false
					case sdl.K_z:
						e.Keyboard[0xA] = false
					case sdl.K_x:
						e.Keyboard[0x0] = false
					case sdl.K_c:
						e.Keyboard[0xB] = false
					case sdl.K_v:
						e.Keyboard[0xF] = false
					}
				} else if et.Type == sdl.KEYDOWN {
					switch et.Keysym.Sym {
					case sdl.K_1:
						e.Keyboard[0x1] = true
					case sdl.K_2:
						e.Keyboard[0x2] = true
					case sdl.K_3:
						e.Keyboard[0x3] = true
					case sdl.K_4:
						e.Keyboard[0xC] = true
					case sdl.K_q:
						e.Keyboard[0x4] = true
					case sdl.K_w:
						e.Keyboard[0x5] = true
					case sdl.K_e:
						e.Keyboard[0x6] = true
					case sdl.K_r:
						e.Keyboard[0xD] = true
					case sdl.K_a:
						e.Keyboard[0x7] = true
					case sdl.K_s:
						e.Keyboard[0x8] = true
					case sdl.K_d:
						e.Keyboard[0x9] = true
					case sdl.K_f:
						e.Keyboard[0xE] = true
					case sdl.K_z:
						e.Keyboard[0xA] = true
					case sdl.K_x:
						e.Keyboard[0x0] = true
					case sdl.K_c:
						e.Keyboard[0xB] = true
					case sdl.K_v:
						e.Keyboard[0xF] = true
					}
				}
			}
		}

		//time.Sleep(int(1.0/speed*1000000.0)*time.Microsecond) // 700 Hz
    //time.Sleep(14280*2 * time.Microsecond) // 35 Hz
		//time.Sleep(14280 * time.Microsecond) // 70 Hz
    time.Sleep(1428 * time.Microsecond) // 700 Hz
		//time.Sleep(1 * time.Second) // 1 Hz
	}

	fmt.Println(e.RAM[0x100:0x10F])

	fmt.Println("Done")

}
