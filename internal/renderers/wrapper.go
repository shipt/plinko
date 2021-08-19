/**
 * Copyright (c) Shipt.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
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
