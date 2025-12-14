package base62

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Encode(n uint64) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte // enough for uint64 in base62
	i := len(b)
	for n > 0 {
		rem := n % 62
		n /= 62
		i--
		b[i] = alphabet[rem]
	}
	return string(b[i:])
}

func Decode(s string) (uint64, bool) {
	var n uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		var v int
		switch {
		case c >= '0' && c <= '9':
			v = int(c - '0')
		case c >= 'A' && c <= 'Z':
			v = 10 + int(c-'A')
		case c >= 'a' && c <= 'z':
			v = 36 + int(c-'a')
		default:
			return 0, false
		}
		n = n*62 + uint64(v)
	}
	return n, true
}
