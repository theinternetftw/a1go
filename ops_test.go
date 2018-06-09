package a1go

import "fmt"
import "testing"

func fmtFlags(flags byte) string {
	out := []byte{'_', '_', '_', '_', '_', '_', '_', '_'}
	if flags&flagNeg != 0 {
		out[0] = byte('N')
	}
	if flags&flagOverflow != 0 {
		out[1] = byte('V')
	}
	if flags&flagOnStack != 0 {
		out[2] = byte('S')
	}
	if flags&flagBrk != 0 {
		out[3] = byte('B')
	}
	if flags&flagDecimal != 0 {
		out[4] = byte('D')
	}
	if flags&flagIrqDisabled != 0 {
		out[5] = byte('I')
	}
	if flags&flagZero != 0 {
		out[6] = byte('Z')
	}
	if flags&flagCarry != 0 {
		out[7] = byte('C')
	}
	return fmt.Sprintf("0x%02x(%s)", flags, out)
}

type decArithTest struct {
	regA       byte
	val        byte
	regP       byte
	result     byte
	resultRegP byte
}

func (d decArithTest) String() string {
	return fmt.Sprintf(
		"0x%02x+0x%02x,C=%v==0x%02x,%s",
		d.regA,
		d.val,
		d.regP&flagCarry == flagCarry,
		d.result,
		fmtFlags(d.resultRegP),
	)
}

func (d decArithTest) runTest(t *testing.T, fnToTest func(cs *cpuState, val byte) byte) {
	cs := cpuState{
		P: d.regP | flagDecimal,
		A: d.regA,
	}
	result := fnToTest(&cs, d.val)
	if result != d.result {
		t.Errorf("got result val of 0x%02x, expected 0x%02x", result, d.result)
	}
	expectedRegP := d.resultRegP | flagDecimal
	if cs.P != expectedRegP {
		t.Errorf("got result regP of %s, expected %s", fmtFlags(cs.P), fmtFlags(expectedRegP))
	}
}

func TestADCDecimalMode(t *testing.T) {
	const c = flagCarry
	adcTests := []decArithTest{
		{0x00, 0x00, 0, 0x00, flagZero},
		{0x79, 0x00, c, 0x80, flagNeg | flagOverflow},
		{0x24, 0x56, 0, 0x80, flagNeg | flagOverflow},
		{0x93, 0x82, 0, 0x75, flagOverflow | flagCarry},
		{0x89, 0x76, 0, 0x65, flagCarry},
		{0x89, 0x76, c, 0x66, flagZero | flagCarry},
		{0x80, 0xf0, 0, 0xd0, flagOverflow | flagCarry},
		{0x80, 0xfa, 0, 0xe0, flagNeg | flagCarry},
		{0x2f, 0x4f, 0, 0x74, 0},
		{0x6f, 0x00, c, 0x76, 0},
	}

	for _, test := range adcTests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			entry := test // make local copy
			//t.Parallel()

			entry.runTest(t, func(cs *cpuState, val byte) byte {
				return cs.adcAndSetFlags(val)
			})
		})
	}
}

func TestSBCDecimalMode(t *testing.T) {
	const c = flagCarry // NOTE: remember carry is inverted for sbc
	sbcTests := []decArithTest{
		{0x00, 0x00, 0, 0x99, flagNeg},
		{0x00, 0x00, c, 0x00, flagZero | flagCarry},
		{0x00, 0x01, c, 0x99, flagNeg},
		{0x0a, 0x00, c, 0x0a, flagCarry},
		{0x0b, 0x00, 0, 0x0a, flagCarry},
		{0x9a, 0x00, c, 0x9a, flagNeg | flagCarry},
		{0x9b, 0x00, 0, 0x9a, flagNeg | flagCarry},
	}

	for _, test := range sbcTests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			entry := test // make local copy
			//t.Parallel()

			entry.runTest(t, func(cs *cpuState, val byte) byte {
				return cs.sbcAndSetFlags(val)
			})
		})
	}
}
