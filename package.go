package rclib

import (
	"errors"
	"log"
)

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

const (
	initial = iota
	uid
	tid
	config
	mesh
	additionalConfig
	data
	checksum
	endbyte
)

func PackageEquals(a, b Package) bool {
	if a.Header != b.Header {
		return false
	}

	if !ConfigEquals(a.Config, b.Config) {
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

func GetUid() byte {
	lastSeenUid++
	return lastSeenUid
}

// Encode returns the encoded byte array of the package.
// TxId should be an identifier unique to each sender/program/component
// If uId is -1, it is set to the last seen UID + 1.
func (p *Package) Encode() []byte {
	var result = []byte{STARTBYTE, byte(p.Header.Uid), p.Header.TransmitterId}
	result = append(result, p.Config.toBytes()...)

	encodedChannelData := make([]byte, dataBytesCount(p.Config.Resolution, p.Config.ChannelCount))
	currentDataByte := 0
	currentDataBit := 0

	for _, channel := range p.Channel {
		for channelBitIndex := 0; channelBitIndex < p.Config.Resolution.BitsPerChannel(); channelBitIndex++ {

			channelBit := byte(channel>>uint(channelBitIndex)) & 1

			encodedChannelData[currentDataByte] |= channelBit << uint(currentDataBit)

			currentDataBit++
			if currentDataBit >= 8 {
				currentDataBit = 0
				currentDataByte++
			}
		}
	}
	result = append(result, encodedChannelData...)
	result = append(result, checkSum(result[1:]))
	result = append(result, ENDBYTE)
	return result
}

// Decode tries to decode the data byte in the context of its current state.
// Returns true, nil when the package is fully decoded.
// Does not produce any useful result when called after package has been decoded.
// Do not do that.
func (p *Package) Decode(dataByte byte) (complete bool, err error) {
	switch p.decodingState {
	case initial: // Initial state
		p.buffer = []byte{}
		if dataByte == STARTBYTE {
			logIfDebug("Found Startbyte.")
			p.decodingState = uid
		}
		break
	case uid: // Start word received
		p.Header.Uid = dataByte
		logIfDebug("Found UID Byte: 0x%X\n", dataByte)
		lastSeenUid = byte(p.Header.Uid)
		p.decodingState = tid
		break
	case tid: // Transmitter Id
		p.Header.TransmitterId = dataByte
		logIfDebug("Found TransmitterID Byte: 0x%X\n", dataByte)
		p.decodingState = config
		break
	case config: // Configuration
		logIfDebug("Found Config Byte: %b\n", dataByte)
		p.Config.Resolution = Resolution(dataByte & 7)
		logIfDebug("Resolution: %d bits per channel\n", p.Config.Resolution.BitsPerChannel())
		p.Config.ChannelCount = []int{1, 2, 4, 8, 16, 32, 64, 256}[dataByte>>3&7]
		logIfDebug("Channel count: %d\n", p.Config.ChannelCount)
		p.totalDataBytesCount = dataBytesCount(p.Config.Resolution, p.Config.ChannelCount)
		logIfDebug("Calculated total Data Bytes: %d\n", p.totalDataBytesCount)
		p.Channel = make([]int, p.Config.ChannelCount)
		p.Config.Error = (dataByte >> 6 & 1) == 1
		logIfDebug("Error bit set: %t\n", p.Config.Error)
		// Following
		if (dataByte >> 7) == 1 {
			logIfDebug("Following bit set. Decoding Mesh Byte next.\n")
			p.decodingState = mesh
		} else {
			logIfDebug("Following bit not set. Decoding Data Byte next.\n")
			p.decodingState = data
		}
		p.decodedDataBytes = 0
		break
	case mesh: // Mesh
		logIfDebug("Found Mesh Byte: %b\n", dataByte)
		p.Config.RoutingLength = int(dataByte & 15 /*0b1111*/)
		logIfDebug("Routing length: %d\n", p.Config.RoutingLength)
		if dataByte>>7 == 1 {
			p.decodingState = additionalConfig
		} else {
			p.decodingState = data
		}
		p.decodingState = 6
		p.decodedDataBytes = 0
		break
	case additionalConfig: // Additional Config
		logIfDebug("Found additional config byte: %b", dataByte)
		p.Config.AdditionalConfig = append(p.Config.AdditionalConfig, dataByte)
		if (dataByte >> 7) == 0 {
			p.decodingState = data
		}
		break
	case data: // Data
		{
			logIfDebug("Found Data Byte %b\n", dataByte)
			for c := 0; c < 8; c++ {
				currentDataBit := 8*p.decodedDataBytes + c
				currentChannel := currentDataBit / p.Config.Resolution.BitsPerChannel()
				if currentChannel >= p.Config.ChannelCount {
					// channelCount * bitsPerChannel < dataBytesCount
					continue
				}
				currentChannelBit := currentDataBit - currentChannel*p.Config.Resolution.BitsPerChannel()
				currentBitValue := (dataByte >> uint(c)) & 1
				logIfDebug("Data bit Nr %d (%d) is bit Nr %d of Channel %d\n", currentDataBit, currentBitValue, currentChannelBit, currentChannel)
				p.Channel[currentChannel] |= int(currentBitValue) << uint(currentChannelBit)
			}
			p.decodedDataBytes = p.decodedDataBytes + 1
			if p.decodedDataBytes >= p.totalDataBytesCount {
				p.decodingState = checksum
			}
		}
		break
	case checksum: // Checksum
		p.checksum = dataByte
		calculatedChecksum := checkSum(p.buffer[1:])
		logIfDebug("Found Checksum: 0x%X\n", dataByte)
		logIfDebug("Calculated Checksum: 0x%X\n", calculatedChecksum)
		if dataByte != calculatedChecksum {
			return true, errors.New("Checksum Incorrect")
		}
		p.decodingState = endbyte
		break
	case endbyte: // End byte
		if dataByte == ENDBYTE {
			logIfDebug("Found End Byte.\n")
			return true, nil
		} else {
			return true, errors.New("No Endbyte")
		}
	default:
		p.decodingState = initial
		break
	}
	p.buffer = append(p.buffer, dataByte)
	return false, nil
}

// DecodePackages is a utility function which decodes all bytes from a channel
// and sends the decoded packages to another channel.
// Errors will be printed to log using log.Println
func DecodePackages(dataIn chan byte, packages chan Package) {
	p := Package{}
	for b := range dataIn {
		finished, err := p.Decode(b)
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

type Header struct {
	Uid           byte
	TransmitterId byte
}
