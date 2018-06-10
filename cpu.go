package a1go

import "fmt"

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

type g6502 struct {
	PC            uint16
	P, A, X, Y, S byte

	IRQ, BRK, NMI, RESET bool
	LastStepsP           byte

	runCycles func(uint)
	write     func(uint16, byte)
	read      func(uint16) byte

	Steps uint64
}

func (cs *g6502) push16(val uint16) {
	cs.push(byte(val >> 8))
	cs.push(byte(val))
}
func (cs *g6502) push(val byte) {
	cs.write(0x100+uint16(cs.S), val)
	cs.S--
}

func (cs *g6502) pop16() uint16 {
	val := uint16(cs.pop())
	val |= uint16(cs.pop()) << 8
	return val
}
func (cs *g6502) pop() byte {
	cs.S++
	result := cs.read(0x100 + uint16(cs.S))
	return result
}

// interrupt info lags behind actual P flag,
// so we need the delay provided by having
// a LastStepsP
func (cs *g6502) interruptsEnabled() bool {
	return cs.LastStepsP&flagIrqDisabled == 0
}

func (cs *g6502) handleInterrupts() {
	if cs.RESET {
		cs.RESET = false
		cs.PC = cs.read16(0xfffc)
		cs.S -= 3
		cs.P |= flagIrqDisabled
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

func (cs *g6502) step() {
	cs.Steps++
	cs.handleInterrupts()
	cs.stepOpcode()
}

func (cs *g6502) read16(addr uint16) uint16 {
	low := uint16(cs.read(addr))
	high := uint16(cs.read(addr + 1))
	return (high << 8) | low
}

func (cs *g6502) write16(addr uint16, val uint16) {
	cs.write(addr, byte(val))
	cs.write(addr+1, byte(val>>8))
}

func (cs *g6502) debugStatusLine() string {
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
}

func (cs *g6502) setOverflowFlag(test bool) {
	if test {
		cs.P |= flagOverflow
	} else {
		cs.P &^= flagOverflow
	}
}

func (cs *g6502) setCarryFlag(test bool) {
	if test {
		cs.P |= flagCarry
	} else {
		cs.P &^= flagCarry
	}
}

func (cs *g6502) setZeroFlag(test bool) {
	if test {
		cs.P |= flagZero
	} else {
		cs.P &^= flagZero
	}
}

func (cs *g6502) setNegFlag(test bool) {
	if test {
		cs.P |= flagNeg
	} else {
		cs.P &^= flagNeg
	}
}

func (cs *g6502) setZeroNeg(val byte) {
	if val == 0 {
		cs.P |= flagZero
	} else {
		cs.P &^= flagZero
	}
	if val&0x80 == 0x80 {
		cs.P |= flagNeg
	} else {
		cs.P &^= flagNeg
	}
}

func (cs *g6502) setNoFlags(val byte) {}
