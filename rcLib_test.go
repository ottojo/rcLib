package rclib

import (
	"encoding/json"
	"reflect"
	"testing"
)

type testCase struct {
	P Package
	B []byte
}

var testCases = []testCase{
	{
		B: []byte{0xc9, 0x1, 0x0, 0x1d, 0x1, 0x8, 0x30, 0x0, 0x1, 0x5, 0x18, 0x70, 0x0, 0x2, 0x4b, 0x93},
		P: Package{
			Header:  Header{1, 0},
			Config:  Configuration{ChannelCount: 8, Resolution: 5},
			Channel: []int{1, 2, 3, 4, 5, 6, 7, 8}}},
	{
		B: []byte{0xc9, 0x1, 0x0, 0x12, 0x1, 0xad, 0xec, 0xf, 0x5c, 0x93},
		P: Package{
			Header:  Header{1, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 2},
			Channel: []int{1, 90, 50, 127}}},
	{
		B: []byte{0xc9, 0x1, 0x0, 0x16, 0x1, 0xd0, 0x82, 0xc, 0xfe, 0x0, 0xb6, 0x93},
		P: Package{
			Header:  Header{1, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 6},
			Channel: []int{1, 90, 50, 127}}},
	{
		B: []byte{0xc9, 0x3, 0x0, 0x92, 0x1, 0x1, 0xad, 0xec, 0xf, 0xdf, 0x93},
		P: Package{
			Header:  Header{3, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 2, RoutingLength: 1},
			Channel: []int{1, 90, 50, 127}}},
	{
		B: []byte{0xc9, 0x3, 0x0, 0xd2, 0x1, 0x1, 0xad, 0xec, 0xf, 0x9f, 0x93},
		P: Package{
			Header:  Header{3, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 2, RoutingLength: 1, Error: true},
			Channel: []int{1, 90, 50, 127}}},
	{
		B: []byte{0xc9, 0x0, 0x0, 0x17, 0xff, 0x4f, 0x6, 0x0, 0xf2, 0xff, 0xac, 0x93},
		P: Package{
			Header:  Header{0, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 7},
			Channel: []int{4095, 100, 512, 4095}}},
	{
		P: Package{
			Header:  Header{0, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 7},
			Channel: []int{4095, 100, 512, 4095}},
		B: []byte{0xc9, 0x0, 0x0, 0x17, 0xff, 0x4f, 0x6, 0x0, 0xf2, 0xff, 0xac, 0x93}},
	{
		P: Package{
			Header:  Header{0, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 7, RoutingLength: 1},
			Channel: []int{4095, 100, 512, 4095}},
		B: []byte{0xc9, 0x0, 0x0, 0x97, 0x1, 0xff, 0x4f, 0x6, 0x0, 0xf2, 0xff, 0x2d, 0x93}},
	{
		P: Package{
			Header:  Header{0, 0},
			Config:  Configuration{ChannelCount: 4, Resolution: 7, RoutingLength: 5},
			Channel: []int{4095, 100, 512, 4095}},
		B: []byte{0xc9, 0x0, 0x0, 0x97, 0x5, 0xff, 0x4f, 0x6, 0x0, 0xf2, 0xff, 0x29, 0x93}},
	{
		P: Package{
			Config:  Configuration{ChannelCount: 4, Resolution: 3},
			Channel: []int{0, 0, 0, 0},
		},
		B: []byte{0xc9, 0x0, 0x0, 0x13, 0x0, 0x0, 0x0, 0x0, 0x13, 0x93},
	}}

func TestPackage_Decode(t *testing.T) {
	for _, tCase := range testCases {
		p := Package{}
		for _, b := range tCase.B {
			f, err := p.Decode(b)
			if err != nil {
				t.Errorf("Got error %v when decoding %v", err, tCase.B)
			}
			if f {
				break
			}
		}

		if !PackageEquals(p, tCase.P) {
			outJ, _ := json.Marshal(p)
			expJ, _ := json.Marshal(tCase.P)
			t.Errorf("From Input %v got Output\n%v, expected \n%v", tCase.B, string(outJ), string(expJ))
		}
	}
}

func TestPackage_Encode(t *testing.T) {
	// TODO more test cases
	for _, tCase := range testCases {
		b := tCase.P.Encode()
		if !reflect.DeepEqual(b, tCase.B) {
			jPackage, _ := json.Marshal(tCase.P)
			t.Errorf("Package %v was encoded to\n%v, expected\n%v", string(jPackage), b, tCase.B)
		}
	}
}
