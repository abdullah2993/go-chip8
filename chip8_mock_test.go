package chip8

import (
	"time"
)

type mockGC struct {
	clearChan chan struct{}
	timeOut   time.Duration
}

var _ GC = (*mockGC)(nil)

func (m *mockGC) Clear() {
	m.clearChan <- struct{}{}
}

func (m *mockGC) Draw(pd [cH][cW]uint8) {
}

func (m *mockGC) isClearPressed() bool {
	select {
	case <-m.clearChan:
		return true
	case <-time.After(m.timeOut):
		return false
	}
}

func newMockGC(timeOut time.Duration) *mockGC {
	return &mockGC{
		clearChan: make(chan struct{}, 1),
		timeOut:   timeOut,
	}
}

type mockSC struct {
	soundChan chan bool
	timeOut   time.Duration
}

var _ SC = (*mockSC)(nil)

func (m *mockSC) Sound(on bool) {

}

func (m *mockSC) isOn() bool {
	select {
	case v := <-m.soundChan:
		return v
	case <-time.After(m.timeOut):
		return false
	}
}

func newMockSC(timeOut time.Duration) *mockSC {
	return &mockSC{
		soundChan: make(chan bool, 1),
		timeOut:   timeOut,
	}
}

type mockIC struct {
}

var _ IC = (*mockIC)(nil)

func (m *mockIC) IsPressed(k uint8) bool {
	return false
}

func (m *mockIC) WaitKey() uint8 {
	return 0
}

func newMockIC() *mockIC {
	return &mockIC{}
}

func uint8SliceEqual(src, dest []uint8) bool {
	if len(src) != len(dest) {
		return false
	}

	for i := 0; i < len(src); i++ {
		if src[i] != dest[i] {
			return false
		}
	}
	return true
}
