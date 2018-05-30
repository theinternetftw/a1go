package a1go

import (
	"fmt"
	"os"
)

type cpuState struct {
	Mem mem

	Screen [240 * 240 * 4]byte

	flipRequested bool
	terminal      terminal

	PC            uint16
	P, A, X, Y, S byte

	IRQ, BRK, NMI, RESET bool
	LastStepsP           byte

	LastKeyState     [256]bool
	NewKeyWasPressed bool
	NewKeyInput      byte

	DebugKeyPressed bool

	NextKeyToDisplay    byte
	ReadyToDisplay      bool
	KeyDisplayRequested bool

	DisplayBeenInitted bool

	Steps  uint64
	Cycles uint64
}

func (cs *cpuState) runCycles(cycles uint) {
	for i := uint(0); i < cycles; i++ {
		cs.Cycles++

		// not great timing, probably
		if cs.KeyDisplayRequested && cs.ReadyToDisplay {
			cs.KeyDisplayRequested = false
			cs.terminal.writeChar(rune(cs.NextKeyToDisplay))
			cs.flipRequested = true
		}
		if !cs.flipRequested {
			cs.ReadyToDisplay = true
		}
	}
}

func (cs *cpuState) debugStatusLine() string {
	if showMemReads {
		fmt.Println()
	}
	opcode := cs.read(cs.PC)
	b2, b3 := cs.read(cs.PC+1), cs.read(cs.PC+2)
	sp := 0x100 + uint16(cs.S)
	s1, s2, s3 := cs.read(sp), cs.read(sp+1), cs.read(sp+2)
	return fmt.Sprintf("Steps: %09d ", cs.Steps) +
		fmt.Sprintf("PC:%04x ", cs.PC) +
		fmt.Sprintf("*PC[:3]:%02x%02x%02x ", opcode, b2, b3) +
		fmt.Sprintf("*S[:3]:%02x%02x%02x ", s1, s2, s3) +
		fmt.Sprintf("opcode:%v ", opcodeNames[opcode]) +
		fmt.Sprintf("A:%02x ", cs.A) +
		fmt.Sprintf("X:%02x ", cs.X) +
		fmt.Sprintf("Y:%02x ", cs.Y) +
		fmt.Sprintf("P:%02x ", cs.P) +
		fmt.Sprintf("S:%02x ", cs.S)
	/*
		return fmt.Sprintf("%04X  ", cs.PC) +
			fmt.Sprintf("%02X %02X %02X  ", opcode, b2, b3) +
			fmt.Sprintf("%v                             ", opcodeNames[opcode]) +
			fmt.Sprintf("A:%02X ", cs.A) +
			fmt.Sprintf("X:%02X ", cs.X) +
			fmt.Sprintf("Y:%02X ", cs.Y) +
			fmt.Sprintf("P:%02X ", cs.P) +
			fmt.Sprintf("SP:%02X", cs.S)
	*/
}

const (
	flagNeg         = 0x80
	flagOverflow    = 0x40
	flagOnStack     = 0x20
	flagBrk         = 0x10
	flagDecimal     = 0x08
	flagIrqDisabled = 0x04
	flagZero        = 0x02
	flagCarry       = 0x01
)

func (cs *cpuState) handleInterrupts() {
	if cs.RESET {
		cs.RESET = false
		cs.PC = cs.read16(0xfffc)
		cs.S -= 3
		cs.P |= flagIrqDisabled
		cs.DisplayBeenInitted = false
	} else if cs.BRK {
		cs.BRK = false
		cs.push16(cs.PC + 1)
		cs.push(cs.P | flagBrk | flagOnStack)
		cs.P |= flagIrqDisabled
		cs.PC = cs.read16(0xfffe)
	} else if cs.NMI {
		cs.NMI = false
		cs.push16(cs.PC)
		cs.push(cs.P | flagOnStack)
		cs.P |= flagIrqDisabled
		cs.PC = cs.read16(0xfffa)
	} else if cs.IRQ {
		cs.IRQ = false
		if cs.interruptsEnabled() {
			cs.push16(cs.PC)
			cs.push(cs.P | flagOnStack)
			cs.P |= flagIrqDisabled
			cs.PC = cs.read16(0xfffe)
		}
	}
	cs.LastStepsP = cs.P
}

func (cs *cpuState) push16(val uint16) {
	cs.push(byte(val >> 8))
	cs.push(byte(val))
}
func (cs *cpuState) push(val byte) {
	cs.write(0x100+uint16(cs.S), val)
	cs.S--
}

func (cs *cpuState) pop16() uint16 {
	val := uint16(cs.pop())
	val |= uint16(cs.pop()) << 8
	return val
}
func (cs *cpuState) pop() byte {
	cs.S++
	result := cs.read(0x100 + uint16(cs.S))
	return result
}

func (cs *cpuState) interruptsEnabled() bool {
	return cs.LastStepsP&flagIrqDisabled == 0
}

const (
	showMemReads  = false
	showMemWrites = false
)

func (cs *cpuState) step() {

	cs.handleInterrupts()

	cs.Steps++

	/*
		// single step w/ debug printout
		{
			if !cs.DebugKeyPressed {
				cs.runCycles(1)
				return
			}
			cs.DebugKeyPressed = false
			fmt.Println(cs.debugStatusLine())
		}
	*/

	cs.stepOpcode()
}

func (cs *cpuState) updateInput(input Input) {
	for i, k := range input.Keys {
		if i > 127 {
			continue
		}

		// for now? lower to upper, ignore real upper
		{
			if i >= 'A' && i <= 'Z' {
				continue
			}
			if i >= 'a' && i <= 'z' {
				i = i - 'a' + 'A'
			}
		}

		if !cs.LastKeyState[i] && k {
			cs.NewKeyInput = byte(i)
			cs.NewKeyWasPressed = true
			cs.DebugKeyPressed = true
		}
		cs.LastKeyState[i] = k
	}
}

func newState(romBytes []byte) *cpuState {
	cs := cpuState{
		Mem:            mem{},
		RESET:          true,
		ReadyToDisplay: true,
	}
	cs.terminal = terminal{
		w:      240,
		h:      192,
		screen: cs.Screen[:],
		font:   a1Font5x7,
	}

	return &cs
}

func emuErr(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}
