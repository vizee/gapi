package ioutil

import (
	"io"
)

func ReadToEnd(r io.Reader, expected int64) ([]byte, error) {
	n := expected
	if n < 0 {
		n = 512
	}

	buf := make([]byte, n)
	i := int64(0)
	for expected < 0 || i < expected {
		if i >= n {
			buf = append(buf, 0)
			n = int64(cap(buf))
			buf = buf[:n]
		}

		nn, err := r.Read(buf[i:n])
		i += int64(nn)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return buf[:i], err
		}
	}
	return buf[:i], nil
}
