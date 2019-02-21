package chip8

import (
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	timeOut, _ := time.ParseDuration("1s")
	gc := newMockGC(timeOut)
	sc := newMockSC(timeOut)
	ic := newMockIC()

	prog := []uint8{0xFE}
	c, err := newChip8(gc, sc, ic, prog)
	if err != nil {
		t.Error(err)
	}

	if c.PC != 0x0200 {
		t.Errorf("Expected pc to be %x got %x", c.PC, 0x0200)
	}

	if !uint8SliceEqual(c.MEM[0:len(fontSet)], fontSet) {
		t.Errorf("Expected font set to be loaded in memory")
	}

	if !uint8SliceEqual(c.MEM[c.PC:int(c.PC)+len(prog)], prog) {
		t.Errorf("Expected program to be loaded in memory")
	}
}

func TestCLS(t *testing.T) {
	timeOut, _ := time.ParseDuration("1s")
	gc := newMockGC(timeOut)
	sc := newMockSC(timeOut)
	ic := newMockIC()

	prog := []uint8{0x00, 0xE0}
	c, err := newChip8(gc, sc, ic, prog)
	if err != nil {
		t.Error(err)
	}
	c.cycle()

	if !gc.isClearPressed() {
		t.Errorf("Expected clear to be called")
	}
}

func TestAnnnAnd6xkkAndFx33(t *testing.T) {
	timeOut, _ := time.ParseDuration("1s")
	gc := newMockGC(timeOut)
	sc := newMockSC(timeOut)
	ic := newMockIC()

	prog := []uint8{0xAF, 0x00, 0x62, 0xFE, 0xF2, 0x33}
	c, err := newChip8(gc, sc, ic, prog)
	if err != nil {
		t.Error(err)
	}

	c.cycle()
	if c.I != 0x0F00 {
		t.FailNow()
	}

	c.cycle()
	if c.V[0x02] != 0xFE {
		t.FailNow()
	}

	c.cycle()
	if c.MEM[0x0F00] != 0x02 {
		t.FailNow()
	}
	if c.MEM[0x0F01] != 0x05 {
		t.FailNow()
	}
	if c.MEM[0x0F02] != 0x04 {
		t.FailNow()
	}
}

func Test1nnn(t *testing.T) {
	timeOut, _ := time.ParseDuration("1s")
	gc := newMockGC(timeOut)
	sc := newMockSC(timeOut)
	ic := newMockIC()

	prog := []uint8{0x12, 0x34}
	c, err := newChip8(gc, sc, ic, prog)
	if err != nil {
		t.Error(err)
	}

	c.cycle()

	if c.PC != 0x0234 {
		t.FailNow()
	}
}
func Test2nnn(t *testing.T) {
	timeOut, _ := time.ParseDuration("1s")
	gc := newMockGC(timeOut)
	sc := newMockSC(timeOut)
	ic := newMockIC()

	prog := []uint8{0x22, 0x34}
	c, err := newChip8(gc, sc, ic, prog)
	if err != nil {
		t.Error(err)
	}

	pc := c.PC

	c.cycle()

	if c.PC != 0x0234 {
		t.Errorf("PC not equal")
	}

	if c.Stack[0] != pc+2 {
		t.Errorf("Stack not equal")
	}
	if c.SP != 1 {
		t.Errorf("Stack pointer not equal")
	}
}
func Test3xkkAnd6xkk(t *testing.T) {
	timeOut, _ := time.ParseDuration("1s")
	gc := newMockGC(timeOut)
	sc := newMockSC(timeOut)
	ic := newMockIC()

	prog := []uint8{0x61, 0x34, 0x31, 0x34, 0x00, 0x00, 0x62, 0x44, 0x32, 0x33}
	c, err := newChip8(gc, sc, ic, prog)
	if err != nil {
		t.Error(err)
	}

	c.cycle()

	if c.V[1] != 0x34 {
		t.Errorf("V1 not loaded")
	}

	pc := c.PC
	c.cycle()

	if c.V[1] != 0x34 {
		t.Errorf("V1 not equal to kk")
	}

	if c.PC != pc+4 {
		t.Errorf("PC not equal")
	}

	c.cycle()
	if c.V[2] != 0x44 {
		t.Errorf("V2 not loaded")
	}

	pc = c.PC
	c.cycle()

	if c.PC != pc+2 {
		t.Errorf("PC not equal")
	}

}

func TestUint8SliceEqual(t *testing.T) {
	tt := []struct {
		src      []uint8
		dest     []uint8
		expected bool
	}{
		{[]uint8{1, 2, 3}, []uint8{1, 2, 3}, true},
		{[]uint8{1, 2, 3, 4}, []uint8{1, 2, 3}, false},
		{[]uint8{1, 2, 3, 4}, []uint8{3, 2, 4, 5, 6}, false},
	}

	for _, tc := range tt {
		if uint8SliceEqual(tc.src, tc.dest) != tc.expected {
			t.Fail()
		}
	}
}
