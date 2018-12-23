package main

import (
	"github.com/theinternetftw/a1go"
	"github.com/theinternetftw/a1go/profiling"
	"github.com/theinternetftw/glimmer"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func main() {

	defer profiling.Start().Stop()

	assert(len(os.Args) == 2 || len(os.Args) == 1, "usage: ./a1go [INPUT_FILENAME]")

	var emu a1go.Emulator

	romFilename := ""
	if len(os.Args) == 2 {
		romFilename = os.Args[1]
		inputBytes, err := ioutil.ReadFile(romFilename)
		dieIf(err)

		emu = a1go.NewEmulatorWithAutokeyInput(inputBytes)
	} else {
		emu = a1go.NewEmulator()
	}

	execPath, err := os.Executable()
	if err == nil {
		execDir := path.Dir(execPath)
		basicPath := path.Join(execDir, "roms", "basic.bin")
		basicBytes, err := ioutil.ReadFile(basicPath)
		if err == nil {
			err := emu.LoadBinaryToMem(0xe000, basicBytes)
			if err != nil {
				dieIf(fmt.Errorf("err when loading basic.bin: %v", err))
			} else {
				fmt.Println("loaded basic!")
			}
		} else {
			fmt.Println("could not auto-load basic:", err)
		}
	} else {
		fmt.Println("could not find executable path:", err)
	}

	screenW := 240
	screenH := 192
	glimmer.InitDisplayLoop("a1go", screenW*2+40, screenH*2+40, screenW, screenH, func(sharedState *glimmer.WindowState) {
		startEmu(sharedState, emu, romFilename)
	})
}

func startEmu(window *glimmer.WindowState, emu a1go.Emulator, romFilename string) {

	// FIXME: settings are for debug right now
	lastVBlankTime := time.Now()

	if romFilename == "" {
		romFilename = "algo"
	}
	snapshotPrefix := romFilename + ".snapshot"
	snapInProgress := false

	numDown := 'x'
	lastNumDown := 'x'
	snapshotMode := 'x'

	for {
		newInput := a1go.Input {}

		hyperMode := false

		window.Mutex.Lock()
		{

			switch {
			case window.CodeIsDown(glimmer.CodeF1):
				newInput.ResetButton = true
			case window.CodeIsDown(glimmer.CodeF2):
				newInput.ClearScreenButton = true
			case window.CodeIsDown(glimmer.CodeF11):
				hyperMode = true
			}

			if window.CodeIsDown(glimmer.CodeF4) {
				snapshotMode = 'm'
			} else if window.CodeIsDown(glimmer.CodeF9) {
				snapshotMode = 'l'
			} else {
				snapInProgress = false
			}

			numDown = 'x'
			for r := '0'; r <= '9'; r++ {
				if window.CharIsDown(r) {
					numDown = r
					break
				}
			}
			if lastNumDown != 'x' {
				if !window.CharIsDown(lastNumDown) {
					lastNumDown = 'x'
				}
			}

			if snapshotMode == 'x' && lastNumDown == 'x' {
				window.CopyKeyCharArray(newInput.Keys[:])
				newInput.Keys['\r'] = window.CodeIsDown(glimmer.CodeReturnEnter)
			}
		}
		window.Mutex.Unlock()

		if numDown > '0' && numDown <= '9' {
			snapFilename := snapshotPrefix+string(numDown)
			if snapshotMode == 'm' {
				if !snapInProgress {
					snapInProgress = true
					lastNumDown = numDown
					snapshotMode = 'x'
					snapshot := emu.MakeSnapshot()
					if len(snapshot) > 0 {
						ioutil.WriteFile(snapFilename, snapshot, os.FileMode(0644))
					}
					fmt.Println("writing snap to", snapFilename)
				}
			} else if snapshotMode == 'l' {
				if !snapInProgress {
					snapInProgress = true
					lastNumDown = numDown
					snapshotMode = 'x'
					snapBytes, err := ioutil.ReadFile(snapFilename)
					if err != nil {
						fmt.Println("failed to load snapshot:", err)
						continue
					}
					newEmu, err := emu.LoadSnapshot(snapBytes)
					if err != nil {
						fmt.Println("failed to load snapshot:", err)
						continue
					}
					emu = newEmu
				}
			}
		}

		emu.UpdateInput(newInput)
		emu.Step()

		if emu.FlipRequested() {
			window.Mutex.Lock()
			copy(window.Pix, emu.Framebuffer())
			window.RequestDraw()
			window.Mutex.Unlock()

			spent := time.Now().Sub(lastVBlankTime)
			toWait := 17*time.Millisecond - spent
			if !hyperMode && toWait > time.Duration(0) {
				<-time.NewTimer(toWait).C
			}
			lastVBlankTime = time.Now()
		}
	}
}

func assert(test bool, msg string) {
	if !test {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func dieIf(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
