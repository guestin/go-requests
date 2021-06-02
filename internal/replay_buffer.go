package internal

import (
	"errors"
	"github.com/guestin/mob/mio"
	"io"
)

type ReplayBuffer struct {
	ReadWriteSeekClose
	raw          io.ReadCloser
	rawReadError error
	eofReach     bool
}

func isEOF(err error) bool {
	return errors.Is(io.EOF, err)
}

func NewReplayBuffer(raw io.ReadCloser) io.ReadSeekCloser {
	return &ReplayBuffer{
		ReadWriteSeekClose: NewSimpleBuffer(),
		raw:                raw,
		rawReadError:       nil,
	}
}

func (this *ReplayBuffer) Read(p []byte) (int, error) {
	if this.eofReach {
		return this.ReadWriteSeekClose.Read(p)
	}
	if this.rawReadError == nil {
		n, err := this.raw.Read(p)
		if err == nil {
			_, _ = this.ReadWriteSeekClose.Write(p[:n])
		} else {
			if isEOF(err) {
				this.eofReach = true
			} else {
				this.rawReadError = err
			}
		}
		return n, err
	}
	return 0, this.rawReadError
}

func (this *ReplayBuffer) Close() error {
	_ = this.ReadWriteSeekClose.Close()
	return this.raw.Close()
}

type ReadWriteSeekClose interface {
	io.ReadSeekCloser
	io.Writer
}

type SimpleBuffer struct {
	data []byte
	rIdx int
}

func NewSimpleBuffer() *SimpleBuffer {
	return &SimpleBuffer{
		data: make([]byte, 0),
		rIdx: 0,
	}
}

func (this *SimpleBuffer) Close() error {
	this.data = make([]byte, 0)
	return nil
}

func (this *SimpleBuffer) Len() int {
	return len(this.data) - this.rIdx
}

func (this *SimpleBuffer) Read(p []byte) (int, error) {
	bufLen := this.Len()
	if bufLen == 0 {
		return 0, io.EOF
	}
	expectN := mio.MinInt(bufLen, len(p))
	copy(p, this.data[this.rIdx:expectN])
	this.rIdx += expectN
	return expectN, nil
}

func (this *SimpleBuffer) Write(p []byte) (n int, err error) {
	this.data = append(this.data, p...)
	return len(p), nil
}

func (this *SimpleBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		this.rIdx += int(offset)
	case io.SeekStart:
		this.rIdx = int(offset)
	case io.SeekEnd:
		this.rIdx = len(this.data) - int(offset)
	}
	return int64(this.rIdx), nil
}

type noOpSeeker struct {
	io.ReadCloser
}

func (this *noOpSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func NoOpSeeker(reader io.ReadCloser) io.ReadSeekCloser {
	return &noOpSeeker{
		reader,
	}
}
