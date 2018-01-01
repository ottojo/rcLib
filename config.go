package rclib

import "encoding/json"

type Configuration struct {
	ChannelCount     int
	Resolution       Resolution
	Error            bool
	RoutingLength    int
	AdditionalConfig []byte
}

func (c *Configuration) IsMeshPackage() bool {
	return c.RoutingLength != 0
}

func ConfigEquals(a, b Configuration) bool {
	if a.ChannelCount != b.ChannelCount {
		return false
	}
	if a.Resolution != b.Resolution {
		return false
	}

	if a.Error != b.Error {
		return false
	}

	if a.RoutingLength != b.RoutingLength {
		return false
	}

	if len(a.AdditionalConfig) != len(b.AdditionalConfig) {
		return false
	}

	for i, v := range a.AdditionalConfig {
		if v != b.AdditionalConfig[i] {
			return false
		}
	}

	return true
}

func (c *Configuration) toBytes() []byte {
	config := []byte{byte(c.Resolution) | channelValue(c.ChannelCount)<<3}
	if c.Error {
		config[0] |= 1 << 6
	}

	if c.IsMeshPackage() || len(c.AdditionalConfig) != 0 {
		config[0] |= 1 << 7
		meshByte := byte(c.RoutingLength)
		config = append(config, meshByte)
	}

	for _, b := range c.AdditionalConfig {
		config[len(config)-1] |= 1 << 7
		config = append(config, b)
	}

	return config
}

type Resolution byte

// Returns resolution value from steps count
// Step count must be one of the following:
// 32, 64, 128, 256, 512, 1024, 2048
func ResolutionFromSteps(steps int) Resolution {
	switch steps {
	case 32:
		return 0
	case 64:
		return 1
	case 128:
		return 2
	case 256:
		return 3
	case 512:
		return 4
	case 1024:
		return 5
	case 2048:
		return 6
	default:
		return 7
	}
}

func (r *Resolution) Steps() int {
	return []int{32, 64, 128, 256, 512, 1024, 2048, 4096}[*r]
}

func (r *Resolution) BitsPerChannel() int {
	return int((*r) + 5)
}

func (r Resolution) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Steps())
}
