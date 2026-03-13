package mem

import "C"
import (
	"C"
	"fmt"
	"rest_demo/zmem/c"
	"unsafe"
)

type Buf struct {
	// 如果存在多个buffer , 则采用链表形式连接起来
	Next *Buf
	// 当前buffer 的缓存容量大小
	Capacity int
	// 当前buffer 的有效数据长度
	length int
	// 未处理数据的头部位置索引
	head int
	// 当前buff 所保存的数据地址
	data unsafe.Pointer
}

func NewBuf(size int) *Buf {
	return &Buf{
		Capacity: size,
		length:   0,
		head:     0,
		Next:     nil,
		data:     c.Malloc(size),
	}
}

func (b *Buf) SetBytes(src []byte) {
	c.Memcpy(unsafe.Pointer(uintptr(b.data)+uintptr(b.head)), src, len(src))
	b.length += len(src)
}

// 获取一个Buf 的数据
func (b *Buf) GetBytes() []byte {
	data := C.GoBytes(unsafe.Pointer(uintptr(b.data)+uintptr(b.head)), C.int(b.length))
	return data
}

// 将其他Buf 对象的数据copy 到自己种
func (b *Buf) Copy(other *Buf) {
	c.Memcpy(b.data, other.GetBytes(), other.length)
	b.head = 0
	b.length = other.length
}

func (b *Buf) Pop(len int) {
	if b.data == nil {
		fmt.Printf("pop data is nil")
		return
	}
	if len > b.length {
		return
	}
	b.length -= len
	b.head += len
}

func (b *Buf) Adjust() {
	if b.head != 0 {
		if b.length != 0 {
			c.Memmove(b.data, unsafe.Pointer(uintptr(b.data)+uintptr(b.head)), b.length)
		}
		b.head = 0
	}
}

func (b *Buf) Clear() {
	b.length = 0
	b.head = 0
}

func (b *Buf) Head() int {
	return b.head
}

func (b *Buf) Length() int {
	return b.length
}
