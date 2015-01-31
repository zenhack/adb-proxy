package common

import (
	"io"
)

func LoudCopy(dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
}
