package renderers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testWriter struct {
	writeFunc func(p []byte) (n int, err error)
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	return w.writeFunc(p)
}

func TestWriteWrapper(t *testing.T) {
	w := testWriter{
		writeFunc: func(p []byte) (n int, err error) {
			return 0, nil
		},
	}
	wrapped := writeWrapper{writer: &w}

	wrapped.write([]byte("test1"))
	assert.Nil(t, wrapped.err)

	w.writeFunc = func(p []byte) (n int, err error) {
		return 0, errors.New("test-error")
	}
	wrapped.write([]byte("test2"))
	assert.NotNil(t, wrapped.err)

	w.writeFunc = func(p []byte) (n int, err error) {
		assert.FailNow(t, "should stop writing after first error")
		return 0, errors.New("test-error")
	}
	wrapped.write([]byte("test3"))
	assert.NotNil(t, wrapped.err)
}
