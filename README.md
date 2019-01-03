# a1go - an apple 1 emulator in go

My other emulators:
[dmgo](https://github.com/theinternetftw/dmgo),
[famigo](https://github.com/theinternetftw/famigo),
[vcsgo](https://github.com/theinternetftw/vcsgo), and
[segmago](https://github.com/theinternetftw/segmago).

#### Features:
 * Only the 6502 monitor is included! It's 1976, $666.66 is a lot of money! You didn't spring for Woz's BASIC!
 * If you have a text file in monitor syntax, put that file in as an argument to have it auto-typed in!
 * Hyperspeed! (hit F11 to speed things up)
 * Quicksave/Quickload, too!
 * Graphical cross-platform support!

That last bit relies on [glimmer](https://github.com/theinternetftw/glimmer). Tested on windows 10 and ubuntu 18.10.

#### Compiling

 * If you have go version >= 1.11, `go build ./cmd/a1go` should be enough.
 * The interested can also see my build script `b` for profiling and such.

#### Important Notes:

 * Reset button is F1
 * Clear Screen in F2
 * Quicksave/Quickload is done by pressing F4 (make quicksave) or F9 (load quicksave), followed by a number key

