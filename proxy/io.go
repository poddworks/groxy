package proxy

import (
	"io"
	"net"
	"time"
)

type CopyIO struct {
	rd time.Duration
	wd time.Duration
}

func (cio *CopyIO) ReadDeadline(d time.Duration) {
	cio.rd = d
}

func (cio *CopyIO) setReadDeadline(r io.Reader, d time.Duration) {
	if d != 0 {
		r.(net.Conn).SetReadDeadline(time.Now().Add(d))
	}
}

func (cio *CopyIO) WriteDeadline(d time.Duration) {
	cio.wd = d
}

func (cio *CopyIO) setWriteDeadline(w io.Writer, d time.Duration) {
	if d != 0 {
		w.(net.Conn).SetWriteDeadline(time.Now().Add(d))
	}
}

func (cio *CopyIO) Copy(dst io.Writer, src io.Reader) (int64, error) {
	return cio.copyBuffer(dst, src, nil)
}

func (cio *CopyIO) CopyBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer")
	}
	return cio.copyBuffer(dst, src, buf)
}

func (cio *CopyIO) copyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
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
