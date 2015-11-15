package proxy

import (
	"io"
	"net"
	"time"
)

// CopyIO extends upon io.Copy but with added support to Read and Write
// deadline
type CopyIO struct {
	rd time.Duration
	wd time.Duration
}

// ReadDeadline sets the duration to deadline from time of Read
func (cio *CopyIO) ReadDeadline(d time.Duration) {
	cio.rd = d
}

func (cio *CopyIO) setReadDeadline(r net.Conn, d time.Duration) {
	if d != 0 {
		r.SetReadDeadline(time.Now().Add(d))
	}
}

// ReadDeadline sets the duration to deadline from time of Write
func (cio *CopyIO) WriteDeadline(d time.Duration) {
	cio.wd = d
}

func (cio *CopyIO) setWriteDeadline(w net.Conn, d time.Duration) {
	if d != 0 {
		w.SetWriteDeadline(time.Now().Add(d))
	}
}

// Copy moves data from src to dst until io.EOF
// Depending on the configuration, each Read and/or Write will be stamped with
// a expected deadline before it should complete
func (cio *CopyIO) Copy(dst net.Conn, src net.Conn) (int64, error) {
	return cio.copyBuffer(dst, src, nil)
}

// CopyBuffer moves data from src to dst until io.EOF using the buffer provided
// by caller
// Depending on the configuration, each Read and/or Write will be stamped with
// a expected deadline before it should complete
func (cio *CopyIO) CopyBuffer(dst net.Conn, src net.Conn, buf []byte) (int64, error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer")
	}
	return cio.copyBuffer(dst, src, buf)
}

func (cio *CopyIO) copyBuffer(dst net.Conn, src net.Conn, buf []byte) (written int64, err error) {
	if buf == nil {
		buf = make([]byte, 32*1024)
	}
	for {
		cio.setReadDeadline(src, cio.rd)
		nr, er := src.Read(buf)
		if nr > 0 {
			cio.setWriteDeadline(dst, cio.wd)
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return written, err
}

func NewCopyIO() *CopyIO {
	return &CopyIO{
		rd: time.Duration(0),
		wd: time.Duration(0),
	}
}
