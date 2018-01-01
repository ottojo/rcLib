package rclib

import (
	"log"
)

const STARTBYTE byte = 0xC9
const ENDBYTE byte = 0x93

var DEBUG = false

var lastSeenUid byte

func channelValue(count int) byte {
	switch count {
	case 1:
		return 0
	case 2:
		return 1
	case 4:
		return 2
	case 8:
		return 3
	case 16:
		return 4
	case 32:
		return 5
	case 64:
		return 6
	default:
		return 7
	}
}

func logIfDebug(format string, a ...interface{}) {
	if DEBUG {
		log.Printf(format, a...)
	}
}

func checkSum(bytes []byte) byte {
	var checksum byte
	for _, v := range bytes {
		checksum ^= v
	}
	return checksum
}

func dataBytesCount(resolution Resolution, channelCount int) int {
	totalBits := resolution.BitsPerChannel() * channelCount
	if totalBits%8 == 0 {
		return totalBits / 8
	} else {
		return int(float64(totalBits)/8.0 + 1)
	}
}

