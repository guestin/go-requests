package internal

import (
	"bytes"
	"fmt"
	"github.com/guestin/mob/mio"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestReplayBuffer_Read(t *testing.T) {
	raw := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	fmt.Println("raw=", mio.ArrToHexStrWithSp(raw, " "))
	oldReader := bytes.NewBuffer(raw)
	replayBuf := NewReplayBuffer(io.NopCloser(oldReader))
	defer mio.CloseIgnoreErr(replayBuf)
	for i := 0; i < 1000; i++ {
		rBuf := make([]byte, 8)
		nRead, err := replayBuf.Read(rBuf)
		if err == io.EOF {
			nSeek, err := replayBuf.Seek(0, io.SeekStart)
			assert.NoError(t, err, "replayBuf seek")
			assert.Equal(t, 0, int(nSeek))
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, len(raw), nRead)
		assert.Equal(t, raw, rBuf, "rBuf!=raw")
		t.Log("test ok")
	}
}
