package stream

import "io"

// fake is used for testing a stream
func fake(rw io.ReadWriter) io.ReadWriter {
	return &stream{
		rw: rw,
	}
}
