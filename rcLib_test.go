package rclib

import (
	"testing"
	"encoding/json"
)

type testCase struct {
	Input  []byte
	Output Package
}

func TestDecode(t *testing.T) {
	testCases := []testCase{
		{
			Input: []byte{0xc9, 0x1, 0x0, 0x1d, 0x1, 0x8, 0x30, 0x0, 0x1, 0x5, 0x18, 0x70, 0x0, 0x2, 0x4b, 0x93},
			Output: Package{
				Header:  Header{1, 0},
				Config:  Configuration{8, 5, false, false, 0},
				Channel: []int{1, 2, 3, 4, 5, 6, 7, 8}}},
		{
			Input: []byte{0xc9, 0x1, 0x0, 0x12, 0x81, 0xb1, 0xec, 0x6f, 0xa0, 0x93},
			Output: Package{
				Header:  Header{1, 0},
				Config:  Configuration{4, 2, false, false, 0},
				Channel: []int{1, 99, 50, 127}}},
		{
			Input: []byte{0xc9, 0x1, 0x0, 0x16, 0x1, 0x18, 0x83, 0xc, 0xfe, 0x60, 0x1f, 0x93},
			Output: Package{
				Header:  Header{1, 0},
				Config:  Configuration{4, 6, false, false, 0},
				Channel: []int{1, 99, 50, 127}}},
		{
			Input: []byte{0xc9, 0x3, 0x0, 0x92, 0x3, 0x81, 0xb1, 0xec, 0x8f, 0xc1, 0x93},
			Output: Package{
				Header:  Header{3, 0},
				Config:  Configuration{4, 2, false, true, 1},
				Channel: []int{1, 99, 50, 127}}},
		{
			Input: []byte{0xc9, 0x0, 0x0, 0x17, 0xff, 0x4f, 0x6, 0x0, 0xf2, 0xff, 0xac, 0x93},
			Output: Package{
				Header:  Header{0, 0},
				Config:  Configuration{4, 7, false, false, 0},
				Channel: []int{4095, 100, 512, 4095}}}}

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
			t.Errorf("From Input %v got Output\n%v, expected \n%v", tCase.Input, string(outJ), string(expJ))
		}
	}
}
