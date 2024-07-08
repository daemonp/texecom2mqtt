package texecom

// CRC8 calculates the 8-bit CRC for Texecom messages
func CRC8(data []byte) byte {
	crc := byte(0)
	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x85
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
