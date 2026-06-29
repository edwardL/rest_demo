package lsm

import (
	"bytes"
)

type Block struct {
	size       int
	entriesCnt int
	buf        bytes.Buffer
}

func NewBlock(conf *Config) *Block {
	return &Block{size: 0}
}

func (b *Block) Append(key, value []byte) {
	b.entriesCnt++
	n := len(key) + len(value)
	b.buf.Write(key)
	b.buf.Write(value)
	b.size += n
}

func (b *Block) Size() int {
	return b.size
}

func (b *Block) FlushTo(buf *bytes.Buffer) (uint64, error) {
	n, err := buf.Write(b.buf.Bytes())
	return uint64(n), err
}

type Filter interface {
	Add(key []byte)
	KeyLen() int
	Hash() []byte
	Reset()
}
