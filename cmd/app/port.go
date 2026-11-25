package main

// port for the given name in the given directory, from a hash of the directory and the name.
// The hash is the POSIX cksum checksum, so the port is the same as from the port function in versitygw.sh.
func port(dir, name string) int {
	b := []byte(dir + "/" + name)

	// cksum appends the length to the data, least significant byte first, without trailing zero bytes.
	for n := len(b); n > 0; n >>= 8 {
		b = append(b, byte(n))
	}

	var crc uint32
	for _, c := range b {
		crc ^= uint32(c) << 24
		for range 8 {
			if crc&0x80000000 != 0 {
				crc = crc<<1 ^ 0x04c11db7
			} else {
				crc <<= 1
			}
		}
	}

	return 20000 + int(^crc%10000)
}
