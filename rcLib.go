package rclib

import (
	"errors"
	"log"
)

const STARTBYTE byte = 0xC9
const ENDBYTE byte = 0x93

var DEBUG = false

var globalUid byte

func PackageEquals(a, b Package) bool {
	if a.Header != b.Header {
		return false
	}
	if a.Config.ChannelCount != b.Config.ChannelCount {
		return false
	}
	if a.Config.Resolution != b.Config.Resolution {
		return false
	}

	if a.Config.Error != b.Config.Error {
		return false
	}

	if a.Config.IsMeshPackage != b.Config.IsMeshPackage {
		return false
	}

	if a.Config.RoutingLength != b.Config.RoutingLength {
		return false
	}

	if len(a.Channel) != len(b.Channel) {
		return false
	}

	for i, v := range a.Channel {
		if v != b.Channel[i] {
			return false
		}
	}
	return true
}

type Package struct {
	Header              Header
	Config              Configuration
	Channel             []int
	buffer              []byte
	decodingState       int
	decodedDataBytes    int
	checksum            byte
	totalDataBytesCount int
}

type Header struct {
	Uid           byte
	TransmitterId byte
}

type Configuration struct {
	ChannelCount  int
	Resolution    Resolution
	Error         bool
	IsMeshPackage bool
	RoutingLength int
	//Additional    []byte
}

type Resolution byte

func logIfDebug(format string, a ...interface{}) {
	if DEBUG {
		log.Printf(format, a...)
	}
}

func (r *Resolution) Steps() int {
	return []int{32, 64, 128, 256, 512, 1024, 2048, 4096}[*r]
}

func (r *Resolution) BitsPerChannel() int {
	return int((*r) + 5)
}

func (p *Package) calcChecksum() byte {
	var checksum byte
	for _, v := range p.buffer[1:] {
		checksum ^= v
	}
	return checksum
}

func DecodePackages(dataIn chan byte, packages chan Package) {
	p := Package{}
	for b := range dataIn {
		finished, err := p.decode(b)
		if err != nil {
			log.Println(err)
			p = Package{}
			continue
		}
		if finished {
			packages <- p
			p = Package{}
		}
	}
	logIfDebug("Data In Channel closed.")
}

func dataBytesCount(resolution Resolution, channelCount int) int {
	totalBits := resolution.BitsPerChannel() * channelCount
	if totalBits%8 == 0 {
		return totalBits / 8
	} else {
		return int(float64(totalBits)/8.0 + 1)
	}
}

func (p *Package) decode(data byte) (bool, error) {
	switch p.decodingState {
	case 0: // Initial state
		p.buffer = []byte{}
		if data == STARTBYTE {
			logIfDebug("Found Startbyte.")
			p.decodingState = 1
		}
		break
	case 1: // Start word received
		p.Header.Uid = data
		logIfDebug("Found UID Byte: 0x%X\n", data)
		globalUid = p.Header.Uid
		p.decodingState = 2
		break
	case 2: // Transmitter Id
		p.Header.TransmitterId = data
		logIfDebug("Found TransmitterID Byte: 0x%X\n", data)
		p.decodingState = 3
		break
	case 3: // Configuration
		logIfDebug("Found Config Byte: %b\n", data)
		p.Config.Resolution = Resolution(data & 7)
		logIfDebug("Resolution: %d bits per channel\n", p.Config.Resolution.BitsPerChannel())
		p.Config.ChannelCount = []int{1, 2, 4, 8, 16, 32, 64, 256}[data>>3&7]
		logIfDebug("Channel count: %d\n", p.Config.ChannelCount)
		p.totalDataBytesCount = dataBytesCount(p.Config.Resolution, p.Config.ChannelCount)
		logIfDebug("Calculated total Data Bytes: %d\n", p.totalDataBytesCount)
		p.Channel = make([]int, p.Config.ChannelCount)
		p.Config.Error = (data >> 6 & 1) == 1
		logIfDebug("Error bit set: %t\n", p.Config.Error)
		// Following
		if (data >> 7) == 1 {
			logIfDebug("Following bit set. Decoding Mesh Byte next.\n")
			p.decodingState = 4
		} else {
			logIfDebug("Following bit not set. Decoding Data Byte next.\n")
			p.decodingState = 5
		}
		p.decodedDataBytes = 0
		break
	case 4: // Mesh
		logIfDebug("Found Mesh Byte: %b\n", data)
		p.Config.IsMeshPackage = (data & 1) == 1
		p.Config.RoutingLength = int((data >> 1) & 15 /*0b1111*/)
		p.decodingState = 5
		p.decodedDataBytes = 0
		break
	case 5: // Data
		{
			logIfDebug("Found Data Byte %b\n", data)
			for c := 0; c < 8; c++ {
				currentDataBit := 8*p.decodedDataBytes + c
				currentChannel := currentDataBit / p.Config.Resolution.BitsPerChannel()
				if currentChannel >= p.Config.ChannelCount {
					// channelCount * bitsPerChannel < dataBytesCount
					continue
				}
				currentChannelBit := currentDataBit - currentChannel*p.Config.Resolution.BitsPerChannel()
				currentBitValue := (data >> uint(c)) & 1
				logIfDebug("Data bit Nr %d (%d) is bit Nr %d of Channel %d\n", currentDataBit, currentBitValue, currentChannelBit, currentChannel)
				p.Channel[currentChannel] |= int(currentBitValue) << uint(currentChannelBit)
			}
			p.decodedDataBytes = p.decodedDataBytes + 1
			if p.decodedDataBytes >= p.totalDataBytesCount {
				p.decodingState = 6
			}
		}
		break
	case 6: // Checksum
		p.checksum = data
		logIfDebug("Found Checksum: 0x%X\n", data)
		logIfDebug("Calculated Checksum: 0x%X\n", p.calcChecksum())
		if data != p.calcChecksum() {
			return true, errors.New("Checksum Incorrect")
		}
		p.decodingState = 7
		break
	case 7: // End byte
		if data == ENDBYTE {
			logIfDebug("Found End Byte.\n")
			return true, nil
		} else {
			return true, errors.New("No Endbyte")
		}
	default:
		p.decodingState = 0
		break
	}
	p.buffer = append(p.buffer, data)
	return false, nil
}
