package bgo

import (
	"errors"
	"io"
	"os"
	"strings"
)

type RESREF struct {
	Name [8]byte
}

func (r *RESREF) String() string {
	str := strings.Split(string(r.Name[0:]), "\x00")[0]
	return str
}

var ErrFormat = errors.New("bgo: Unknown file format")

type format struct {
	name, magic string
	decode      func(io.ReadSeeker) (BgoFile, error)
}

type BgoFile interface {
	WriteJson(io.WriteSeeker) error
}

var formats []format

func sniff(r io.ReadSeeker) format {
	var b [8]uint8
	amt, err := r.Read(b[0:])
	r.Seek(int64(-1*amt), os.SEEK_CUR)
	if err != nil {
		return format{}
	}

	for _, f := range formats {
		if match(f.magic, b[0:]) {
			return f
		}
	}
	return format{}
}

func match(magic string, b []byte) bool {
	if len(magic) != len(b) {
		return false
	}

	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

func RegisterFormat(name, magic string, decode func(io.ReadSeeker) (BgoFile, error)) {
	formats = append(formats, format{name, magic, decode})
}

func Open(r io.ReadSeeker) (BgoFile, error) {
	f := sniff(r)
	if f.decode == nil {
		return nil, ErrFormat
	}

	file, err := f.decode(r)

	return file, err
}
