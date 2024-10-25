package chip8

import (
  "os"
  "log"
  "fmt"
  "crypto/rand"
)


type Emulator struct {

  RAM [4096]byte
  Registers [16]byte
  PC uint16
  SP uint16
  I uint16
  VF uint16
  DelayTimer byte
  SoundTimer byte
  Display [32][64]byte
  Keyboard [16]bool
  Draw bool
  Stack [16]uint16
  Instructions uint64
}


func LoadRom(e *Emulator, filename string){
  file, err := os.Open(filename)
  if err != nil {
    log.Fatal(err)
  }
  defer file.Close()

  fileInfo, err := file.Stat()
  if err != nil {
    log.Fatal(err)
  }

  fileSize := fileInfo.Size()

  if fileSize > 4096-512{
    log.Fatal("To large file, won't fit in RAM")
  }

  file.Read(e.RAM[512:])
}

type Instruction struct{
  Full uint16
  Addr uint16
  Nibble byte
  X byte
  Y byte
  KK byte
}

func parseInstruction(high byte, low byte) *Instruction{
  i := &Instruction{}
  i.Full = uint16(high) << 8 + uint16(low)
  i.Addr = i.Full & 0x7FF
  i.Nibble = low & 0xF
  i.X = high & 0x0F
  i.Y = low >> 4
  i.KK = low

  return i
}

func (e *Emulator) PrintRAM(){
  last := len(e.RAM) - 1
  for ; last > 0; last--{
    if e.RAM[last] != 0{
      break
    }
  }
  width := 16

  for i, b := range e.RAM[:last]{
    if i % width == 0{
      fmt.Println()
      fmt.Printf("%04X: ", i)
    }
    fmt.Printf("%02X ", b)
  }
  fmt.Println()
}

func (e *Emulator) Tick() {
  in := parseInstruction(e.RAM[e.PC], e.RAM[e.PC+1])
  log.Println()

  e.Instructions ++

  log.Printf("%d, PC: 0x%04X, SP: 0x%03X, Full: 0x%04X, Addr: 0x%03X, nibble: 0x%X, x:0x%02X, y:0x%02X, kk:0x%02X", e.Instructions, e.PC, e.SP, in.Full, in.Addr, in.Nibble, in.X, in.Y, in.KK)
	log.Println(e.Registers)

  /*
  regs := "Registers - "
  for i, v := range e.Registers[:8]{
    regs += fmt.Sprintf("%d-0x%X, ", i, v)
  }
  log.Println(regs)
  regs = "Registers - "
  for i, v := range e.Registers[8:]{
    regs += fmt.Sprintf("%d-0x%X, ", i+8, v)
  }
  log.Println(regs)
  */

  incPC := true

  // Allways increment the PC
  defer func() {
    if incPC{
      e.PC += 2
    }
  }()

  // Count down the timers
  if e.DelayTimer > 0{
    e.DelayTimer --
  }

  if e.SoundTimer > 0{
    e.SoundTimer --
  }

  switch {

  // Jump. This is ignored in modern interpreters
  /*
  case in.Full & 0xF000 == 0x0000:
    log.Printf("0x0nnn 'Jump' to 0x%04X", in.Addr)
    break
    e.PC = in.Addr
    incPC = false
  */

  // CLS, Clear screen
  case in.Full == 0x00E0:
    log.Println("Running 'Clear Screen'")
    e.Display = [32][64]byte{}
    e.Draw = true

  // Ret
  case in.Full == 0x00EE:
    log.Printf("Running 'Ret'. Returning to 0x%X", e.Stack[e.SP])
    e.PC = e.Stack[e.SP]
    e.SP --

  // JP addr
  case in.Full & 0xF000 == 0x1000:
    log.Printf("Running 'Jump' to 0x%X", in.Addr)
    e.PC = in.Addr
    // Skip incrementing the PC
    incPC = false

  // Call addr
  case in.Full & 0xF000 == 0x2000:
    log.Printf("Running 'Call' 0x%X", in.Addr)
    e.SP ++
    e.Stack[e.SP] = e.PC
    e.PC = in.Addr
    incPC = false

  // SE Vx, byte
  case in.Full & 0xF000 == 0x3000: // TODO: Probs wrong
    // Skip next instruction if register x == kk
    if e.Registers[in.X] == in.KK {
      e.PC += 2
      log.Printf("Running skip equals, skipping")
    } else{
      log.Printf("Running skip equals, not skipping")
    }

  // SNE Vx, byte
  case in.Full & 0xF000 == 0x4000:
    log.Printf("Running skip not equals")
    // Skip next instruction if register x != kk
    if e.Registers[in.X] != in.KK {
      e.PC += 2
    }

  // SE Vx, Vy
  case in.Full & 0xF00F == 0x5000:
    log.Printf("Running skip equals Vx Vy")
    // Skip next instruction if register x != kk
    if e.Registers[in.X] == e.Registers[in.Y] {
      e.PC += 2
    }

  // LD Vx, byte
  case in.Full & 0xF000 == 0x6000:
    log.Printf("Storing KK (0x%X) in Vx (%d)", in.KK, in.X)
    e.Registers[in.X] = in.KK

  // ADD Vx, byte
  case in.Full & 0xF000 == 0x7000:
    log.Printf("Adding kk (%d) to Vx (#%d: %d -> %d)", in.KK, in.X, e.Registers[in.X], e.Registers[in.X] + in.KK)
    e.Registers[in.X] += in.KK

  // LD Vx, Vy
  case in.Full & 0xF00F == 0x8000:
    log.Printf("Storing Vy in Vx")
    e.Registers[in.X] = e.Registers[in.Y]

  // OR Vx, Vy
  case in.Full & 0xF00F == 0x8001:
    log.Printf("Bitwise 'or' of Vy and Vx, Store in Vx")
    e.Registers[in.X] = e.Registers[in.X] | e.Registers[in.Y]

  // AND Vx, Vy
  case in.Full & 0xF00F == 0x8002:
    log.Printf("Bitwise 'and' of Vy and Vx, Store in Vx")
    e.Registers[in.X] = e.Registers[in.X] & e.Registers[in.Y]

  // XOR Vx, Vy
  case in.Full & 0xF00F == 0x8003:
    log.Printf("Bitwise 'xor' of Vy and Vx, Store in Vx")
    e.Registers[in.X] = e.Registers[in.X] ^ e.Registers[in.Y]

  // ADD Vx, Vy
  case in.Full & 0xF00F == 0x8004:
    log.Printf("ADD Vy and Vx, Store in Vx")
    res := uint16(e.Registers[in.X]) + uint16(e.Registers[in.Y]) 

    if res > 255 {
      // Add carry bit to e.VF
      e.VF = 1
    } else {
      e.VF = 0
    }
    e.Registers[in.X] = byte(res)

  // SUB Vx, Vy
  case in.Full & 0xF00F == 0x8005:
    log.Printf("SUB Vy and Vx, Store in Vx")

    if e.Registers[in.X] > e.Registers[in.Y] {
      // Add borrow bit to e.VF
      e.VF = 1
    } else {
      e.VF = 0
    }
    e.Registers[in.X] = e.Registers[in.X] - e.Registers[in.Y]

  // SHR Vx{, Vy}
  case in.Full & 0xF00F == 0x8006:
    log.Printf("Shift Vx right, store in Vx")

    if e.Registers[in.X] & 0b00000001 == 1{
      e.VF = 1
    } else {
      e.VF = 0
    }
    e.Registers[in.X] = e.Registers[in.X] >> 1


  // SUBN Vx, Vy
  case in.Full & 0xF00F == 0x8007:
    log.Printf("SUB Vy and Vx, Store in Vx")

    if e.Registers[in.Y] > e.Registers[in.X] {
      // Add borrow bit to e.VF
      e.VF = 1
    } else {
      e.VF = 0
    }
    e.Registers[in.X] = e.Registers[in.Y] - e.Registers[in.X]


  // SHL Vx{, Vy}
  case in.Full & 0xF00F == 0x800E:
    log.Printf("Shift Vx left, store in Vx")

    if e.Registers[in.X] & 0b10000000 == 0b10000000{
      e.VF = 1
    } else {
      e.VF = 0
    }
    e.Registers[in.X] = e.Registers[in.X] << 1


  // SNE Vx, Vy
  case in.Full & 0xF00F == 0x9000:
    log.Printf("Skip next instruction if Vx != Vy")
    if e.Registers[in.X] != e.Registers[in.X]{
      e.PC += 2
    }

  // LD I, addr
  case in.Full & 0xF000 == 0xA000:
    log.Printf("Setting register I to 0X%x", in.Addr)
    e.I = in.Addr


  // JP V0, addr
  case in.Full & 0xF000 == 0xB000:
    log.Printf("Jumping to V0 (0X%x) + 0X%x, = 0X%x", e.Registers[0], in.Addr, e.Registers[0] + byte(in.Addr))
    // Jump, don't increment PC after this
    incPC = false

    e.PC = uint16(e.Registers[0] + byte(in.Addr))


  // RND Vx, byte
  case in.Full & 0xF000 == 0xC000:

    _, err := rand.Read(e.Registers[in.X:in.X+1])
    if err != nil {
      log.Fatal(err)
    }
    log.Printf("Generated random value 0x%X and stored in register %d", e.Registers[in.X], in.X)

  // DRW Vx, Vy, nibble
  case in.Full & 0xF000 == 0xD000:
    log.Println("Drawing a sprite on screen")
    log.Printf("Start Coordinates - x: %d, y: %d", e.Registers[in.X], e.Registers[in.Y])
    
    // Signal that we should re-draw the screen
    e.Draw = true

    for y, line := range e.RAM[e.I: e.I + uint16(in.Nibble)]{
      bits := [8]byte{}
      y=y

      for i, _ := range bits{
        bits[i] = line & 128
        line = line << 1
      }

      for x, bit := range bits{
        xx := (x + int(e.Registers[in.X])) % 64
        yy := (y + int(e.Registers[in.Y])) % 32
        if bit > 0 && e.Display[yy][xx] > 0{
          e.VF = 1
        }

        if bit > 0 && e.Display[yy][xx] > 0{
          // Turn off pixel
          e.Display[yy][xx] = 0
        } else if bit > 1 && e.Display[yy][xx] == 0{
          // Turn on pixel
          e.Display[yy][xx] = 1
        }
      } 
    }

  // SKP Vx
  case in.Full & 0xF0FF == 0xE09E:
    if e.Keyboard[e.Registers[in.X]] == true{
      e.PC += 2
    }

  // SKNP Vx
  case in.Full & 0xF0FF == 0xE0A1:
    if e.Keyboard[e.Registers[in.X]] == false{
      e.PC += 2
    }

  // LD Vx, DT
  case in.Full & 0xF0FF == 0xF007:
    e.Registers[in.X] = e.DelayTimer

  // LD Vx, K
  case in.Full & 0xF0FF == 0xF00A: // TODO: Wait for keypress and store value in Vx when key is pressed
    log.Println("Waiting for keypress")
    pressed := false

    for i, status := range e.Keyboard{
      if status == true{
        pressed = true
        e.Registers[in.X] = byte(i)
        log.Printf("%X was pressed", i)
      }
    }
    
    if pressed == false{
      incPC = false
    }

  // LD DT, Vx, Set delay timer
  case in.Full & 0xF0FF == 0xF015:
    e.DelayTimer = e.Registers[in.X]

  // LD ST, Vx, Set Sound timer
  case in.Full & 0xF0FF == 0xF018:
    e.SoundTimer = e.Registers[in.X]

  // ADD I, Vx
  case in.Full & 0xF0FF == 0xF01E:
    e.I += uint16(e.Registers[in.X])

  // LD F, Vx. Load memory location for sprite into I
  case in.Full & 0xF0FF == 0xF029:
    e.I = uint16(in.X*5)

  // LD B, Vx
  case in.Full & 0xF0FF == 0xF033:
    hundreds := e.Registers[in.X] / 100
    tens := e.Registers[in.X] / 10
    singles := e.Registers[in.X] % 10

    e.RAM[e.I] = hundreds
    e.RAM[e.I+1] = tens
    e.RAM[e.I+2] = singles

  // LD [I], Vx. Load register 0..x into RAM at location I
  case in.Full & 0xF0FF == 0xF055:
    
    for i, b := range e.Registers[:in.X]{
      loc := e.I + uint16(i)
      log.Printf("Storing byte: 0x%X in location 0x%X", b, loc)
      e.RAM[loc] = b

    }
    
    //copy(e.RAM[e.I:], e.Registers[:in.X]) // TODO: Check if this should be +1?

  // LD Vx, [I]. Load RAM at i into registers 0..x
  case in.Full & 0xF0FF == 0xF065:
    copy(e.Registers[:], e.RAM[e.I: e.I+uint16(in.X)+1])
    log.Println(e.RAM[e.I: e.I+uint16(in.X)+1])


  default:
    log.Fatal("Unknown instruction")
  }

}

func New() *Emulator{

  e := &Emulator{}

  // Set PC to start of program
  e.PC = 0x200

  // Draw initial screen imediately
  e.Draw = true

  // Set up RAM
  sprites := []byte {
    // Zero
    0b11110000,
    0b10010000,
    0b10010000,
    0b10010000,
    0b11110000,

    // One
    0b00100000, 
    0b01100000,
    0b00100000,
    0b00100000,
    0b01110000,

    // Two
    0b11110000,
    0b00010000,
    0b11110000,
    0b10000000,
    0b11110000,

    // Three
    0b11110000,
    0b00010000,
    0b11110000,
    0b00010000,
    0b11110000,

    // Four
    0b10010000,
    0b10010000,
    0b11110000,
    0b00010000,
    0b00010000,

    // Five
    0b11110000,
    0b10000000,
    0b11110000,
    0b00010000,
    0b11110000,

    // Six
    0b11110000,
    0b10000000,
    0b11110000,
    0b10010000,
    0b11110000,

    // Seven
    0b11110000,
    0b00010000,
    0b00100000,
    0b01000000,
    0b01000000,

    // Eight
    0b11110000,
    0b10010000,
    0b11110000,
    0b10010000,
    0b11110000,

    // Nine
    0b11110000,
    0b10010000,
    0b11110000,
    0b00010000,
    0b11110000,

    // A
    0b11110000,
    0b10010000,
    0b11110000,
    0b10010000,
    0b10010000,

    // B
    0b11100000,
    0b10010000,
    0b11100000,
    0b10010000,
    0b11100000,

    // C
    0b11110000,
    0b10000000,
    0b10000000,
    0b10000000,
    0b11110000,

    // D
    0b11100000,
    0b10010000,
    0b10010000,
    0b10010000,
    0b11100000,

    // E
    0b11110000,
    0b10000000,
    0b11110000,
    0b10000000,
    0b11110000,

    // F
    0b11110000,
    0b10000000,
    0b11110000,
    0b10000000,
    0b10000000,

  }
  copy(e.RAM[0:], sprites)

  return e
}
