package main

import (
	"github.com/theinternetftw/a1go"
	"github.com/theinternetftw/a1go/profiling"
	"github.com/theinternetftw/a1go/platform"

	"golang.org/x/mobile/event/key"

	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {

	defer profiling.Start().Stop()

	//assert(len(os.Args) == 2, "usage: ./a1go ROM_FILENAME")
	//romFilename := os.Args[1]

	//romBytes, err := ioutil.ReadFile(romFilename)
	//dieIf(err)

	emu := a1go.NewEmulator([]byte{})

	screenW := 240
	screenH := 192
	platform.InitDisplayLoop(screenW*2+40, screenH*2+40, screenW, screenH, func(sharedState *platform.WindowState) {
		startEmu(sharedState, emu)
	})
}

func startEmu(window *platform.WindowState, emu a1go.Emulator) {

	// FIXME: settings are for debug right now
	lastVBlankTime := time.Now()

	snapshotPrefix := "a1go" + ".snapshot"


	for {
		newInput := a1go.Input {}
		snapshotMode := 'x'
		numDown := 'x'

		window.Mutex.Lock()
		{
			window.CopyKeyCharArray(newInput.Keys[:])
			if window.CodeIsDown(key.CodeF1) {
				newInput.ResetButton = true
			}
			if window.CodeIsDown(key.CodeF2) {
				newInput.ClearScreenButton = true
			}
			/*
			for r := '0'; r <= '9'; r++ {
				if window.CharIsDown(r) {
					numDown = r
					break
				}
			}
			if window.CharIsDown('m') {
				snapshotMode = 'm'
			} else if window.CharIsDown('l') {
				snapshotMode = 'l'
			}
			*/
		}
		window.Mutex.Unlock()

		if numDown > '0' && numDown <= '9' {
			snapFilename := snapshotPrefix+string(numDown)
			if snapshotMode == 'm' {
				snapshotMode = 'x'
				snapshot := emu.MakeSnapshot()
				if len(snapshot) > 0 {
					ioutil.WriteFile(snapFilename, snapshot, os.FileMode(0644))
				}
			} else if snapshotMode == 'l' {
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

		emu.UpdateInput(newInput)
		emu.Step()

		if emu.FlipRequested() {
			window.Mutex.Lock()
			copy(window.Pix, emu.Framebuffer())
			window.RequestDraw()
			window.Mutex.Unlock()

			spent := time.Now().Sub(lastVBlankTime)
			toWait := 17*time.Millisecond - spent
			if toWait > time.Duration(0) {
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
