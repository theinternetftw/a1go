package a1go

// Emulator exposes the public facing fns for an emulation session
type Emulator interface {
	Step()

	MakeSnapshot() []byte
	LoadSnapshot([]byte) (Emulator, error)

	Framebuffer() []byte
	FlipRequested() bool

	UpdateInput(input Input)
}

// Input covers all outside info sent to the Emulator
// TODO: add dt?
type Input struct {
	// Keys is a bool array of keydown state
	Keys              [256]bool
	ResetButton       bool
	ClearScreenButton bool
}

func (emu *emuState) UpdateInput(input Input) {
	emu.updateInput(input)
}

// NewEmulator creates an emulation session
func NewEmulator() Emulator {
	return newState()
}

// NewEmulatorWithAutokeyInput creates an emulation session with input to be autokeyed in from the start
func NewEmulatorWithAutokeyInput(input []byte) Emulator {
	return newStateWithAutokeyInput(input)
}

func (emu *emuState) MakeSnapshot() []byte {
	return emu.makeSnapshot()
}

func (emu *emuState) LoadSnapshot(snapBytes []byte) (Emulator, error) {
	return emu.loadSnapshot(snapBytes)
}

// Framebuffer returns the current state of the screen
func (emu *emuState) Framebuffer() []byte {
	return emu.framebuffer()
}

// FlipRequested indicates if a draw request is pending
// and clears it before returning
func (emu *emuState) FlipRequested() bool {
	return emu.flipRequested()
}

func (emu *emuState) Step() {
	emu.step()
}
