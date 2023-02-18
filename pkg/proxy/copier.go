package proxy

import "io"

// Copier 数据转发器，扩展后
type Copier interface {
	// Copy 将数据从 src 复制到 dst
	Copy(dst io.Writer, src io.Reader) (int64, error)
}

type SimpleCopier struct{}

func (c *SimpleCopier) Copy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}
