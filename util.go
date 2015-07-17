package goduino

func from7Bit(b0 byte, b1 byte) byte {
	return (b0 & 0x7F) | ((b1 & 0x7F) << 7)
}

func to7Bit(i byte) []byte {
	return []byte{i & 0x7f, (i >> 7) & 0x7f}
}

func intto7Bit(i int) []byte {
	return []byte{byte(i & 0x7f), byte((i >> 7) & 0x7f), byte((i >> 14) & 0x7f)}
}

func multibyteString(data []byte) (str string) {

	if len(data)%2 != 0 {
		data = append(data, 0)
	}

	for i := 0; i < len(data); i = i + 2 {
		str = str + string(from7Bit(data[i], data[i+1]))
	}

	return
}
