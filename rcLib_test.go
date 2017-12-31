package rclib

import (
	"encoding/json"
	"testing"
)

type testCase struct {
	Input  []byte
	Output Package
}

func TestDecode(t *testing.T) {
	testCases := []testCase{{
		Input: []byte{0xc9, 0x1, 0x0, 0x1d, 0x1, 0x8, 0x30, 0x0, 0x1, 0x5, 0x18, 0x70, 0x0, 0x2, 0x4b, 0x93},
		Output: Package{
			Header:  Header{1, 0},
			Config:  Configuration{8, 5, false, false, 0, 0},
			Channel: []uint16{1, 2, 3, 4, 5, 6, 7, 8}}}}

	for _, tCase := range testCases {
		p := Package{}
		for _, b := range tCase.Input {
			f, err := p.decode(b)
			if err != nil {
				t.Errorf("Got error %v when decoding %v", err, tCase.Input)
			}
			if f {
				break
			}
		}

		if !PackageEquals(p, tCase.Output) {
			outJ, _ := json.Marshal(p)
			expJ, _ := json.Marshal(tCase.Output)
			t.Errorf("From Input %v got Output %v, expected %v", tCase.Input, string(outJ), string(expJ))
		}
	}
}
