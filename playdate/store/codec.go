package store

const (
	headerSize       = 16
	maximumStoreSize = 16 * 1024 * 1024
)

var magic = [4]byte{'G', 'P', 'D', 'S'}

func encode(version uint32, payload []byte) []byte {
	encoded := make([]byte, headerSize+len(payload))
	copy(encoded[:4], magic[:])
	putUint32(encoded[4:8], version)
	putUint32(encoded[8:12], uint32(len(payload)))
	putUint32(encoded[12:16], checksum(payload))
	copy(encoded[headerSize:], payload)
	return encoded
}

func decode(encoded []byte, maximumSize uint32) (uint32, []byte, error) {
	if len(encoded) < headerSize || string(encoded[:4]) != string(magic[:]) {
		return 0, nil, ErrCorrupt
	}
	version := uint32At(encoded[4:8])
	size := uint32At(encoded[8:12])
	if version == 0 || size > maximumSize {
		return 0, nil, ErrCorrupt
	}
	if uint64(size)+headerSize != uint64(len(encoded)) {
		return 0, nil, ErrCorrupt
	}
	payload := append([]byte(nil), encoded[headerSize:]...)
	if uint32At(encoded[12:16]) != checksum(payload) {
		return 0, nil, ErrCorrupt
	}
	return version, payload, nil
}

func putUint32(target []byte, value uint32) {
	target[0] = byte(value)
	target[1] = byte(value >> 8)
	target[2] = byte(value >> 16)
	target[3] = byte(value >> 24)
}

func uint32At(source []byte) uint32 {
	return uint32(source[0]) |
		uint32(source[1])<<8 |
		uint32(source[2])<<16 |
		uint32(source[3])<<24
}

func checksum(payload []byte) uint32 {
	value := uint32(2166136261)
	for _, octet := range payload {
		value ^= uint32(octet)
		value *= 16777619
	}
	return value
}
