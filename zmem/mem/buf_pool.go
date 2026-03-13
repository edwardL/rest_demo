package mem

import "sync"

const (
	m4K   = 4096
	m16K  = 16 * 1024
	m32K  = 32 * 1024
	m256K = 256 * 1024
	m1M   = 1048576
	m4M   = 4 * 1024 * 1024
	m8M   = 8 * 1024 * 1024
)

type Pool map[int]*Buf

type BufPool struct {
	Pool     Pool
	PoolLock sync.RWMutex

	TotalMem int64
}

var bufPoolInstance *BufPool
var once sync.Once

func MemPool() *BufPool {
	once.Do(func() {
		bufPoolInstance = new(BufPool)
		bufPoolInstance.Pool = make(Pool)
		bufPoolInstance.TotalMem = 0
	})
	return bufPoolInstance
}
