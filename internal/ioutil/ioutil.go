package ioutil

import (
	"io"
)

func ReadLimited(r io.Reader, n int64, limit int64) ([]byte, error) {
	const initSize = 512
	if n < initSize {
		n = initSize
	}
	if n > limit {
		n = limit
	}
	buf := make([]byte, n)
	i := int64(0)
	for i < limit {
		if i >= n {
			n = n + n
			if n > limit {
				n = limit
			}
			newbuf := make([]byte, n)
			copy(newbuf, buf)
			buf = newbuf
		}

		nn, err := r.Read(buf[i:n])
		i += int64(nn)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return buf, err
		}
	}
	return buf, nil
}
