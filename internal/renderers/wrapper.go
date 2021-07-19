// Copyright 2021, Shipt. All rights reserved.
// Licensed under the Apache License
package renderers

import (
	"io"
)

type writeWrapper struct {
	writer io.Writer
	err    error
}

func (w *writeWrapper) write(p []byte) {
	if w.err != nil {
		return
	}

	_, w.err = w.writer.Write(p)
}
