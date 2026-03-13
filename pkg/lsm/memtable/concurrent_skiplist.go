package memtable

import (
	"sync"
	"sync/atomic"
)

// 基于节点粒度锁实现的并发安全的跳表结构
type ConcurrentSkipList struct {
	//当前跳表中存在的元素个数，通过 atomic.Int32 保证增减操作的原子性
	cap atomic.Int32

	// 跳表的头结节点
	head *node
	// 对象池 复用跳表中创建和删除的 node 结构，减轻 gc 压力
	nodesCache sync.Pool

	// 比较 node key 大小的规则，倘若 key1 < key2 返回 true，否则返回 false
	compareFunc func(key1, key2 any) bool

	DeleteMutex sync.RWMutex
}

type node struct {
	key, val any
	// nexts[i] 为第 i 层的 next 节点
	nexts []*node
	// 每个节点持有一把节点粒度的读写锁，后续可以作为左边界锁
	sync.RWMutex
}

// 构造一个并发安全的跳表，需要注入比较 key 大小的规则函数
func NewConcurrentSkipList(compareFunc func(key1, key2 any) bool) *ConcurrentSkipList {
	return &ConcurrentSkipList{
		head: &node{
			nexts: make([]*node, 1),
		},
		// 初始化 node 对象池
		nodesCache: sync.Pool{
			New: func() any {
				return &node{}
			},
		},
		// 注入 key 比较函数
		compareFunc: compareFunc,
	}
}

func (c *ConcurrentSkipList) Get(key any) (any, bool) {
	c.DeleteMutex.RLock()
	defer c.DeleteMutex.RUnlock()

	// 沿着头节点 head 从最高层出发进行检索
	move := c.head
	// 通过 last 记录上一层所在的节点位置，避免在下降的过程中对于同一个节点反复加多次左边界节点锁
	var last *node
	for level := len(c.head.nexts) - 1; level >= 0; level-- {
		// 在同一层中一路无锁穿越，直到来到左边界
		for move.nexts[level] != nil && c.compareFunc(move.nexts[level].key, key) {
			move = move.nexts[level]
		}
		// 走到左边界
		// // 通过 last 指针保证对同一个节点只会加一次左边界节点锁
		if move != last {
			// get 操作针对左边界节点加读锁，保证多个 get 操作可以共享
			move.RLock()
			defer move.RUnlock()
			// 更新 last 指针引用
			last = move
		}
		// 倘若找到目标，则直接返回
		if move.nexts[level] != nil && move.nexts[level].key == key {
			return move.nexts[level].val, true
		}
	}
	// 遍历完也没找到目标，说明 key 不存在
	return 0, false
}

func (c *ConcurrentSkipList) search(key any) *node {
	// 沿着头节点 head 从最高层出发进行检索
	move := c.head
	// 通过 last 记录上一层所在的节点位置，避免在下降的过程中对于同一个节点反复加多次左边界节点锁
	var last *node
	for level := len(c.head.nexts) - 1; level >= 0; level-- {
		// 在同一层中一路无锁穿越，直到来到左边界
		for move.nexts[level] != nil && c.compareFunc(move.nexts[level].key, key) {
			move = move.nexts[level]
		}
		// 走到左边界
		// // 通过 last 指针保证对同一个节点只会加一次左边界节点锁
		if move != last {
			move.RLock()
			defer move.RUnlock()
			// 更新 last 指针引用
			last = move
		}
		// 倘若找到目标，则直接返回
		if move.nexts[level] != nil && move.nexts[level].key == key {
			return move.nexts[level]
		}
	}
	return nil
}

// 写入key-value 对
func (c *ConcurrentSkipList) Put(key, val any) {
	// 取 deleteMutex 的读锁. 和 delete 操作实现互斥，但是 get 操作和 put 操作本身可以共享，基于更细粒度的节点锁实现互斥

}
