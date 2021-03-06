package a1go

import "fmt"

const ramBank1Size = 0xc000
const ramBank2Size = 0x1000

type mem struct {
	RAMBank1 [ramBank1Size]byte
	RAMBank2 [ramBank2Size]byte
}

var monitorROM = [256]byte{
	0xD8, 0x58, 0xA0, 0x7F, 0x8C, 0x12, 0xD0, 0xA9,
	0xA7, 0x8D, 0x11, 0xD0, 0x8D, 0x13, 0xD0, 0xC9,
	0xDF, 0xF0, 0x13, 0xC9, 0x9B, 0xF0, 0x03, 0xC8,
	0x10, 0x0F, 0xA9, 0xDC, 0x20, 0xEF, 0xFF, 0xA9,
	0x8D, 0x20, 0xEF, 0xFF, 0xA0, 0x01, 0x88, 0x30,
	0xF6, 0xAD, 0x11, 0xD0, 0x10, 0xFB, 0xAD, 0x10,
	0xD0, 0x99, 0x00, 0x02, 0x20, 0xEF, 0xFF, 0xC9,
	0x8D, 0xD0, 0xD4, 0xA0, 0xFF, 0xA9, 0x00, 0xAA,
	0x0A, 0x85, 0x2B, 0xC8, 0xB9, 0x00, 0x02, 0xC9,
	0x8D, 0xF0, 0xD4, 0xC9, 0xAE, 0x90, 0xF4, 0xF0,
	0xF0, 0xC9, 0xBA, 0xF0, 0xEB, 0xC9, 0xD2, 0xF0,
	0x3B, 0x86, 0x28, 0x86, 0x29, 0x84, 0x2A, 0xB9,
	0x00, 0x02, 0x49, 0xB0, 0xC9, 0x0A, 0x90, 0x06,
	0x69, 0x88, 0xC9, 0xFA, 0x90, 0x11, 0x0A, 0x0A,
	0x0A, 0x0A, 0xA2, 0x04, 0x0A, 0x26, 0x28, 0x26,
	0x29, 0xCA, 0xD0, 0xF8, 0xC8, 0xD0, 0xE0, 0xC4,
	0x2A, 0xF0, 0x97, 0x24, 0x2B, 0x50, 0x10, 0xA5,
	0x28, 0x81, 0x26, 0xE6, 0x26, 0xD0, 0xB5, 0xE6,
	0x27, 0x4C, 0x44, 0xFF, 0x6C, 0x24, 0x00, 0x30,
	0x2B, 0xA2, 0x02, 0xB5, 0x27, 0x95, 0x25, 0x95,
	0x23, 0xCA, 0xD0, 0xF7, 0xD0, 0x14, 0xA9, 0x8D,
	0x20, 0xEF, 0xFF, 0xA5, 0x25, 0x20, 0xDC, 0xFF,
	0xA5, 0x24, 0x20, 0xDC, 0xFF, 0xA9, 0xBA, 0x20,
	0xEF, 0xFF, 0xA9, 0xA0, 0x20, 0xEF, 0xFF, 0xA1,
	0x24, 0x20, 0xDC, 0xFF, 0x86, 0x2B, 0xA5, 0x24,
	0xC5, 0x28, 0xA5, 0x25, 0xE5, 0x29, 0xB0, 0xC1,
	0xE6, 0x24, 0xD0, 0x02, 0xE6, 0x25, 0xA5, 0x24,
	0x29, 0x07, 0x10, 0xC8, 0x48, 0x4A, 0x4A, 0x4A,
	0x4A, 0x20, 0xE5, 0xFF, 0x68, 0x29, 0x0F, 0x09,
	0xB0, 0xC9, 0xBA, 0x90, 0x02, 0x69, 0x06, 0x2C,
	0x12, 0xD0, 0x30, 0xFB, 0x8D, 0x12, 0xD0, 0x60,
	0x00, 0x00, 0x00, 0x0F, 0x00, 0xFF, 0x00, 0x00,
}

func (emu *emuState) read(addr uint16) byte {
	var val byte
	switch {
	case addr < ramBank1Size:
		val = emu.Mem.RAMBank1[addr]

	case addr == 0xd010:
		val = 0x80 | emu.NewKeyInput
		emu.NewKeyWasPressed = false
	case addr == 0xd011:
		val = boolBit(emu.NewKeyWasPressed, 7)
	case addr == 0xd012:
		val = boolBit(!emu.ReadyToDisplay, 7) | emu.NextKeyToDisplay

	case addr >= 0xe000 && addr < 0xf000:
		val = emu.Mem.RAMBank2[addr-0xe000]

	case addr >= 0xf000 && addr < 0xff00:
		val = 0xff // unused ROM
	case addr >= 0xff00:
		val = monitorROM[addr-0xff00]
	default:
		emuErr(fmt.Sprintf("unimplemented read: 0x%04x", addr))
	}
	if showMemReads {
		fmt.Printf("read(0x%04x) = 0x%02x\n", addr, val)
	}
	return val
}

func (emu *emuState) write(addr uint16, val byte) {
	switch {
	case addr < ramBank1Size:
		emu.Mem.RAMBank1[addr] = val

	case addr == 0xd011:
		// ctrl for PIA setup after RESET, ignored here
	case addr == 0xd012:
		if emu.DisplayBeenInitted {
			emu.NextKeyToDisplay = val & 0x7f
			emu.KeyDisplayRequested = true
			emu.ReadyToDisplay = false
		} else {
			if val == 0x7f {
				emu.DisplayBeenInitted = true
			}
		}
		// fmt.Println("display:", val&0x7f)
	case addr == 0xd013:
		// ctrl for PIA setup after RESET, ignored here

	case addr >= 0xe000 && addr < 0xf000:
		emu.Mem.RAMBank2[addr-0xe000] = val

	case addr >= 0xf000:
		// nop, this is ROM

	default:
		emuErr(fmt.Sprintf("unimplemented: write(0x%04x, 0x%02x)", addr, val))
	}
	if showMemWrites {
		fmt.Printf("write(0x%04x, 0x%02x)\n", addr, val)
	}
}
