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

func (cs *cpuState) UpdateInput(input Input) {
	cs.updateInput(input)
}

// NewEmulator creates an emulation session
func NewEmulator(rom []byte) Emulator {
	return newState(rom)
}

func (cs *cpuState) MakeSnapshot() []byte {
	return cs.makeSnapshot()
}

func (cs *cpuState) LoadSnapshot(snapBytes []byte) (Emulator, error) {
	return cs.loadSnapshot(snapBytes)
}

// Framebuffer returns the current state of the screen
func (cs *cpuState) Framebuffer() []byte {
	return cs.framebuffer()
}

// FlipRequested indicates if a draw request is pending
// and clears it before returning
func (cs *cpuState) FlipRequested() bool {
	return cs.flipRequested()
}

func (cs *cpuState) Step() {
	cs.step()
}
