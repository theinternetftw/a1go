package a1go

import (
	"github.com/theinternetftw/cpugo/virt6502"

	"fmt"
	"os"
)

type emuState struct {
	Mem mem

	CPU virt6502.Virt6502

	Screen [240 * 240 * 4]byte

	Terminal terminal

	autokeyInput []byte

	LastKeyState     [256]bool
	NewKeyWasPressed bool
	NewKeyInput      byte

	DebugKeyPressed bool

	NextKeyToDisplay    byte
	ReadyToDisplay      bool
	KeyDisplayRequested bool

	DisplayBeenInitted bool

	Cycles       uint64
	FrameCounter uint64
}

const clocksPerFrame = 14318100 / 14 / 60

func (emu *emuState) flipRequested() bool {
	result := emu.Terminal.flipRequested
	emu.Terminal.flipRequested = false
	return result
}

func (emu *emuState) framebuffer() []byte {
	return emu.Terminal.screen
}

func (emu *emuState) runCycles(cycles uint) {

	emu.Cycles += uint64(cycles)

	// not great timing, probably
	if emu.KeyDisplayRequested && emu.ReadyToDisplay {
		emu.KeyDisplayRequested = false
		emu.Terminal.writeChar(rune(emu.NextKeyToDisplay))
	}
	if !emu.Terminal.flipRequested {
		emu.ReadyToDisplay = true
	}

	emu.FrameCounter += uint64(cycles)
	if emu.FrameCounter >= clocksPerFrame {
		emu.FrameCounter = 0
		emu.Terminal.flipRequested = true
	}
}

const (
	showMemReads  = false
	showMemWrites = false
)

func (emu *emuState) step() {

	/*
		// single step w/ debug printout
		{
			if !cs.DebugKeyPressed {
				cs.runCycles(1)
				return
			}
			cs.DebugKeyPressed = false
			if showMemReads {
				fmt.Println()
			}
			fmt.Println(emu.CPU.debugStatusLine())
		}
	*/

	emu.CPU.Step()
}

func (emu *emuState) updateInput(input Input) {

	if len(emu.autokeyInput) > 0 {
		if !emu.NewKeyWasPressed {
			input.Keys[emu.autokeyInput[0]] = true
			emu.autokeyInput = emu.autokeyInput[1:]
		}
	}

	// convert lower to upper case
	for i := 0; i < 26; i++ {
		cap := 'A' + i
		low := 'a' + i
		input.Keys[cap] = input.Keys[cap] || input.Keys[low]
		input.Keys[low] = false
	}

	// apple1 monitor expects underscore as a "rubout" / pseudo-backspace key
	//
	// interesting thought here: https://www.applefritter.com/content/how-get-backspace-rubout-working-apple-i-clone-smc-kr3600-keyboard-encoder
	//
	// maybe woz made a mistake and thought the underscore character that
	// appeared on some keyboards when you hit DEL (0x7f) was just the regular
	// underscore. Which it wasn't, apparently.
	//
	// TODO: decide if this is the right thing to do, or just have people
	// figure out that they have to type underscores to rubout chars on the
	// monitor (not backspace, as the video term can't back up).
	//
	// e.g. this might break programs that expect you to be able to type \b's.
	//
	if input.Keys[8] {
		input.Keys[8] = false
		input.Keys[0x5f] = true
	}

	for i, down := range input.Keys {
		if i > 127 {
			continue
		}

		if !emu.LastKeyState[i] && down {
			emu.NewKeyInput = byte(i)
			emu.NewKeyWasPressed = true
			emu.DebugKeyPressed = true
		}
		emu.LastKeyState[i] = down
	}

	if input.ResetButton {
		emu.reset()
	}
	if input.ClearScreenButton {
		// looking at real apple1 demos, I think
		// this is the real behavior...
		emu.Terminal.newline()
	}
}

func (emu *emuState) loadBinaryToMem(addr uint16, bin []byte) error {
	if len(bin)+int(addr) > 0x10000 {
		return fmt.Errorf("binary len %v too big to load at %v", len(bin), addr)
	}
	for i, b := range bin {
		i16 := uint16(i)
		emu.write(addr+i16, b)
	}
	return nil
}

func (emu *emuState) reset() {
	emu.DisplayBeenInitted = false
	emu.Terminal.clearScreen()
	emu.CPU.RESET = true
}

func newStateWithAutokeyInput(input []byte) *emuState {
	emu := newState()
	emu.autokeyInput = input
	return emu
}

func makeTerminal(emu *emuState) terminal {
	return terminal{
		W:      240,
		H:      192,
		screen: emu.Screen[:],
		font:   a1Font5x7,
	}
}
func unpackTerminalFromSnap(emu *emuState) {
	emu.Terminal.flipRequested = true
	emu.Terminal.screen = emu.Screen[:]
	emu.Terminal.font = a1Font5x7
}

func newState() *emuState {
	emu := emuState{
		Mem:            mem{},
		ReadyToDisplay: true,
	}
	emu.CPU = virt6502.Virt6502{
		RESET:     true,
		RunCycles: emu.runCycles,
		Write:     emu.write,
		Read:      emu.read,
		Err:       func(e error) { emuErr(e) },
	}
	emu.Terminal = makeTerminal(&emu)

	return &emu
}

func emuErr(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}
