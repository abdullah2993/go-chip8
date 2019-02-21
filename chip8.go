package chip8

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"runtime/debug"
	"time"
)

const memSize = 0x0FFF
const stackSize = 0xFF
const regs = 16
const pixels = 64 * 32 // Width: 64, Height: 32
const cW = 64
const cH = 32

var fontSet = []uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type chip8 struct {
	MEM   [memSize]uint8
	V     [regs]uint8
	I     uint16
	PC    uint16
	Stack [stackSize]uint16
	SP    uint16
	DT    uint8
	ST    uint8
	gc    GC
	sc    SC
	ic    IC
	gfx   [cH][cW]uint8
}

func (c *chip8) Loop() (stop func()) {
	t := time.NewTicker((100 / 6) * time.Millisecond)
	go func() {
		defer func() {
			fmt.Printf("PANIC: %s", debug.Stack())
			if err := recover(); err != nil {
				fmt.Printf("Recovered: %v", err)
				dump(c)
			}
		}()
		for range t.C {
			c.cycle()
		}
	}()
	return t.Stop
}

func (c *chip8) cycle() {
	ins := uint16(c.MEM[c.PC])<<8 | uint16(c.MEM[c.PC+1])
	c.PC += 2
	if ins == 0x00E0 { // CLS
		for i := 0; i < cH; i++ {
			for j := 0; j < cW; j++ {
				c.gfx[i][j] = 0
			}
		}

		c.gc.Clear()
	} else if ins == 0x00EE { // RET
		c.SP--
		c.PC = c.Stack[c.SP]
	} else {
		switch ins & 0xF000 {
		case 0x0000: // fix?
			c.PC = ins & 0x0FFF
		case 0x1000: // JP addr
			c.PC = ins & 0x0FFF
		case 0x2000: // CALL addr
			c.Stack[c.SP] = c.PC
			c.SP++
			c.PC = ins & 0x0FFF
		case 0x3000: // SE Vx, byte
			x := (ins & 0x0F00) >> 8
			kk := ins & 0x00FF
			if c.V[x] == uint8(kk) {
				c.PC += 2
			}
		case 0x4000: // SNE Vx, byte
			x := (ins & 0x0F00) >> 8
			kk := ins & 0x00FF
			if c.V[x] != uint8(kk) {
				c.PC += 2
			}
		case 0x5000: // SE Vx, Vy
			x := (ins & 0x0F00) >> 8
			y := (ins & 0x00F0) >> 4
			if c.V[x] == c.V[y] {
				c.PC += 2
			}
		case 0x6000: // LD Vx, byte
			x := (ins & 0x0F00) >> 8
			kk := ins & 0x00FF
			c.V[x] = uint8(kk)
		case 0x7000: // ADD Vx, byte
			x := (ins & 0x0F00) >> 8
			kk := ins & 0x00FF
			c.V[x] += uint8(kk)
		case 0x8000:
			op := ins & 0x000F
			x := (ins & 0x0F00) >> 8
			y := (ins & 0x00F0) >> 4
			switch op {
			case 0x00: // LD Vx, Vy
				c.V[x] = c.V[y]
			case 0x01: // OR Vx, Vy
				c.V[x] |= c.V[y]
			case 0x02: // AND Vx, Vy
				c.V[x] &= c.V[y]
			case 0x03: // XOR Vx, Vy
				c.V[x] ^= c.V[y]
			case 0x04: // ADD Vx, Vy
				r := uint16(c.V[x]) + uint16(c.V[y])
				c.V[x] = uint8(r & 0x00FF)
				c.V[0x0F] = 0
				if r > 0xFF {
					c.V[0x0F] = 1
				}
			case 0x05: // SUB Vx, Vy
				c.V[0x0F] = 0
				if c.V[x] > c.V[y] {
					c.V[0x0F] = 1
				}
				c.V[x] -= c.V[y]
			case 0x06: // SHR Vx {, Vy}
				c.V[0x0F] = c.V[x] & 0x01
				c.V[x] >>= 2
			case 0x07: // SUBN Vx, Vy
				c.V[0x0F] = 0
				if c.V[y] > c.V[x] {
					c.V[0x0F] = 1
				}
				c.V[x] = c.V[y] - c.V[x]
			case 0x0E: // SHL Vx {, Vy}
				c.V[0x0F] = c.V[x] >> 7
				c.V[x] <<= 2
			}

		case 0x9000: // SNE Vx, Vy
			x := (ins & 0x0F00) >> 8
			y := (ins & 0x00F0) >> 4
			if c.V[x] != c.V[y] {
				c.PC += 2
			}
		case 0xA000: // Annn - LD I, addr
			c.I = ins & 0x0FFF
		case 0xB000: // Bnnn - JP V0, addr
			c.PC = (ins & 0x0FFF) + uint16(c.V[0])
		case 0xC000: // RND Vx, byte
			x := (ins & 0x0F00) >> 8
			rnd := uint8(rand.Int31n(256))
			kk := ins & 0x00FF
			c.V[x] = rnd & uint8(kk)
		case 0xD000: // DRW Vx, Vy, nibble
			x := (ins & 0x0F00) >> 8
			y := (ins & 0x00F0) >> 4
			n := (ins & 0x000F)
			xo := c.V[x]
			yo := uint16(c.V[y])
			col := false
			for i := uint16(0); i < n; i++ {
				line := c.MEM[c.I+i]
				ay := (i + yo) % cH
				for j := uint8(0); j < 8; j++ {
					ax := (j + xo) % cW
					val := c.gfx[ay][ax]
					sval := (line >> (7 - j)) & 0x01
					c.gfx[ay][ax] ^= sval
					if val == 1 && sval == 1 {
						col = true
					}
				}
			}
			c.V[0x0F] = 0
			if col {
				c.V[0x0F] = 1
			}
			c.gc.Draw(c.gfx)
		case 0xE000:
			x := (ins & 0x0F00) >> 8
			switch ins & 0x00FF {
			case 0x9E:
				if c.ic.IsPressed(c.V[x]) {
					c.PC += 2
				}
			case 0xA1:
				if !c.ic.IsPressed(c.V[x]) {
					c.PC += 2
				}
			}
		case 0xF000:
			x := (ins & 0x0F00) >> 8
			switch ins & 0x00FF {
			case 0x0007: // LD Vx, DT
				c.V[x] = c.DT
			case 0x000A: // LD Vx, K
				c.V[x] = c.ic.WaitKey()
			case 0x0015: // LD DT, Vx
				c.DT = c.V[x]
			case 0x0018: // LD ST, Vx
				c.ST = c.V[x]
			case 0x001E: // ADD I, Vx
				c.I += uint16(c.V[x])
			case 0x0029: // LD F, Vx
				c.I = uint16(c.V[x] * 5)
			case 0x0033: // LD B, Vx
				c.MEM[c.I] = (c.V[x] / 100) % 10
				c.MEM[c.I+1] = (c.V[x] / 10) % 10
				c.MEM[c.I+2] = c.V[x] % 10
			case 0x0055: // LD [I], Vx
				copy(c.MEM[c.I:], c.V[:x+1]) //weird huh? :D
			case 0x0065: // LD Vx, [I]
				copy(c.V[:], c.MEM[c.I:c.I+x+1]) //weird huh? :D

			}
		default:
			panic(fmt.Sprintf("Invalid instruction found: %0000X", ins))
		}

	}

	c.decrementTimers()
}

func (c *chip8) decrementTimers() {
	if c.DT > 0 {
		c.DT--
	}

	if c.ST > 0 {
		c.ST--
		if c.ST == 0 {
			c.sc.Sound(false)
		} else {
			c.sc.Sound(true)
		}
	}
}

// Chip8 returns a chip8 CPU
type Chip8 interface {
	Loop() (stop func())
}

// NewChip8 creates a new Chip8 CPU
func NewChip8(gc GC, sc SC, ic IC, prog []uint8) (Chip8, error) {
	return newChip8(gc, sc, ic, prog)
}

// NewChip8FromROM creates a new Chip8 CPU and loads the ROM from file
func NewChip8FromROM(gc GC, sc SC, ic IC, filename string) (Chip8, error) {
	prog, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to load rom: %v", err)
	}
	return newChip8(gc, sc, ic, prog)
}

func newChip8(gc GC, sc SC, ic IC, prog []uint8) (*chip8, error) {
	if gc == nil {
		return nil, fmt.Errorf("Invalid graphics controller")

	}

	if sc == nil {
		return nil, fmt.Errorf("Invalid sound controller")

	}

	if ic == nil {
		return nil, fmt.Errorf("Invalid input controller")

	}

	if prog == nil || len(prog) > memSize-0x200 {
		return nil, fmt.Errorf("Invalid program")
	}
	c := new(chip8)
	c.gc = gc
	c.sc = sc
	c.ic = ic
	copy(c.MEM[:], fontSet)

	copy(c.MEM[0x0200:], prog)

	c.PC = 0x0200
	return c, nil
}

// GC is the graphic controller
type GC interface {
	Clear()
	Draw(pd [cH][cW]uint8)
}

// SC is the souond cotroller
type SC interface {
	Sound(on bool)
}

// IC is the input cotroller
type IC interface {
	IsPressed(k uint8) bool
	WaitKey() uint8
}

func dump(c *chip8) {
	fmt.Printf("\n\n------DUMP START------\n\n")
	fmt.Printf("PC:\t%04X\n", c.PC)
	fmt.Printf("I:\t%04X\n", c.I)
	fmt.Printf("SP:\t%04X\n", c.SP)
	fmt.Printf("DT:\t%02X\n", c.DT)
	fmt.Printf("ST:\t%02X\n", c.ST)
	fmt.Printf("Registers: \n")
	for i := 0; i < regs; i++ {
		fmt.Printf("\tV%x:\t%02X\n", i, c.V[i])
	}
	fmt.Printf("Stack: \n")
	for i := uint16(0); i < stackSize; i++ {
		fmt.Printf("\tStack[%x]:\t%02X\n", i, c.Stack[i])
	}
	start := uint16(0)
	if c.PC > 20 {
		start = c.PC - 20
	}
	fmt.Printf("MEM: \n")
	for i := start; i <= c.PC; i += 2 {
		fmt.Printf("\t%04X\t%04X\n", i, uint16(c.MEM[i])<<8|uint16(c.MEM[i+1]))
	}
	fmt.Printf("\n\n------DUMP END------\n\n")
}
