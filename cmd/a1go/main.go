package main

import (
	"github.com/theinternetftw/a1go"
	"github.com/theinternetftw/a1go/profiling"
	"github.com/theinternetftw/glimmer"

	"fmt"
	"io/ioutil"
	"os"
	"path"
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
	glimmer.InitDisplayLoop(glimmer.InitDisplayLoopOptions{
		WindowTitle:  "a1go",
		WindowWidth:  screenW*2 + 40,
		WindowHeight: screenH*2 + 40,
		RenderWidth:  screenW,
		RenderHeight: screenH,
		InitCallback: func(sharedState *glimmer.WindowState) {
			startEmu(sharedState, emu, romFilename)
		},
	})
}

func startEmu(window *glimmer.WindowState, emu a1go.Emulator, romFilename string) {

	frameTimer := glimmer.MakeFrameTimer()

	if romFilename == "" {
		romFilename = "algo"
	}
	snapshotPrefix := romFilename + ".snapshot"
	snapInProgress := false

	numDown := 'x'
	lastNumDown := 'x'
	snapshotMode := 'x'

	for {
		newInput := a1go.Input{}

		hyperMode := false

		window.InputMutex.Lock()
		{

			switch {
			case window.CodeIsDown(glimmer.KeyCodeF1):
				newInput.ResetButton = true
			case window.CodeIsDown(glimmer.KeyCodeF2):
				newInput.ClearScreenButton = true
			case window.CodeIsDown(glimmer.KeyCodeF11):
				hyperMode = true
			}

			if window.CodeIsDown(glimmer.KeyCodeF4) {
				snapshotMode = 'm'
			} else if window.CodeIsDown(glimmer.KeyCodeF9) {
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
				newInput.Keys['\r'] = window.CodeIsDown(glimmer.KeyCodeEnter)
			}
		}
		window.InputMutex.Unlock()

		if numDown > '0' && numDown <= '9' {
			snapFilename := snapshotPrefix + string(numDown)
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
			frameTimer.MarkRenderComplete()
			window.RenderMutex.Lock()
			copy(window.Pix, emu.Framebuffer())
			window.RenderMutex.Unlock()

			if !hyperMode {
				<-window.DrawNotifier
			}
			frameTimer.MarkFrameComplete()
			frameTimer.PrintStatsEveryXFrames(60 * 5)
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
