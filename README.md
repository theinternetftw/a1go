# a1go - an apple 1 emulator in go

My other emulators:
[dmgo](https://github.com/theinternetftw/dmgo),
[famigo](https://github.com/theinternetftw/famigo),
[vcsgo](https://github.com/theinternetftw/vcsgo), and
[segmago](https://github.com/theinternetftw/segmago).

#### Features:
 * Only the 6502 monitor is included! It's 1976, and you didn't spring for the tape/BASIC upgrade!
 * If you have a text file in monitor syntax, put that file in as an argument to have it auto-typed in!
 * Hyperspeed! (hit F11 to speed things up)
 * Quicksave/Quickload, too!
 * Graphical cross-platform support!

#### Dependencies:

 * You can compile on windows with no C dependencies.
 * Other platforms should do whatever the [ebiten](https://github.com/hajimehoshi/ebiten) page says, which is what's currently under the hood.

#### Compiling

 * If you have go version >= 1.18, `go build ./cmd/a1go` should be enough.
 * The interested can also see my build script `b` for profiling and such.
 * Non-windows users will need ebiten's dependencies.

#### Important Notes:

 * Reset button is F1
 * Clear Screen in F2
 * Quicksave/Quickload is done by pressing F4 (make quicksave) or F9 (load quicksave), followed by a number key

