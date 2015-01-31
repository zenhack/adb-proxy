package common

import (
	"bufio"
	"io"

	"fmt"
	"os"
)

var logfile io.Writer

func log(fs string, args ...interface{}) {
	return
	if logfile == nil {
		logfile, _ = os.Create("/data/local/tmp/logfile")
	}
	if logfile == nil {
		logfile, _ = os.Create("/tmp/logfile")
	}
	fmt.Fprintf(logfile, fs, args...)
}

//type Encoding []byte

const (
	//ServerEncoding = Encoding([]byte("0123456789abcdef"))
	//ClientEncoding = Encoding([]byte("ghijklmnopqrstuv"))
	ServerStart = 'a'
	ClientStart = 'A'
)

func terminal(start byte) byte {
	return start + ('z' - 'a')
}

type Decoder struct {
	r io.Reader
	enc byte
}

func NewDecoder(r io.Reader, enc byte) *Decoder {
	return &Decoder{r: r, enc: enc}
}

func (d *Decoder) readByte() (byte, error) {
	buf := []byte{0}
	n, err := d.r.Read(buf)
	if n == 0 && err == nil {
		panic("zero read!")
	}
	return buf[0], err
}

func (d *Decoder) Read(p []byte) (n int, err error) {
	n = 0
	for i := range(p) {

		msn, err := d.readNibble()
		if err != nil {
			return n, err
		}
		if msn == 0xff {
			break
		}
		lsn, err := d.readNibble()
		if err != nil {
			return n, err
		}
		if lsn == 0xff {
			panic("lsn terminator!")
		}
		n++
		p[i] = (msn << 4) | lsn
		log("Read Byte: %c\n", p[i])
	}
	log("Read: %q", p)
	return n, nil
}

func (d *Decoder) readNibble() (byte, error) {
	c, err := d.readByte()
//	log("Read Nibble: %c\n", c)
	if err != nil {
		return 0, err
	}
	value := c - d.enc
	for err == nil && value > 0xf {
		// sentinel value. ick ick.
		if c == terminal(d.enc) {
			return 0xff, nil
		}
		c, err = d.readByte()
//		log("Read Nibble: %c\n", c)
		value = c - d.enc
//		log("Nibble Value: %d\n", value)
	}
	if err != nil {
		return 0, err
	}
//	log("Final Nibble Value: %d\n", value)
	return value, err
}

type Encoder struct {
	w *bufio.Writer
	enc byte
}

func NewEncoder(w io.Writer, enc byte) *Encoder {
	return &Encoder{w: bufio.NewWriter(w), enc: enc}
}

func (e *Encoder) Write(p []byte) (n int, err error) {
	n = 0
	linechar := 0
	for i := range(p) {
		msn := (p[i] >> 4) & 0xf
		lsn := p[i] & 0xf

		msn += e.enc
		lsn += e.enc

		err := e.w.WriteByte(msn)
		if err != nil {
			return n, err
		}
		err = e.w.WriteByte(lsn)
		if err != nil {
			return n, err
		}
		n++
		linechar += 2
		if linechar > 72 {
			err = e.w.WriteByte('\n')
			if err != nil {
				return n, err
			}
			linechar = 0
		}
	}
	err = e.w.WriteByte(terminal(e.enc))
	if err != nil {
		return n, err
	}
	err = e.w.WriteByte('\n')
	if err != nil {
		return n, err
	}
	return n, e.w.Flush()
}
