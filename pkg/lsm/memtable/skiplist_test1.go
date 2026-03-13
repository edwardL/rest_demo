package memtable

// import "math/rand"

// type node struct {
// 	nexts    []*node
// 	key, val int
// }
// type SkipList struct {
// 	head *node
// }

// func (s *SkipList) search(key int) *node {
// 	// 每次检索从头部出发
// 	move := s.head
// 	for level := len(s.head.nexts) - 1; level >= 0; level-- {
// 		// 在每一层中持续向右遍历，直到下一个节点不存在或者 key 值大于等于 key
// 		for move.nexts[level] != nil && move.nexts[level].key < key {
// 			move = move.nexts[level]
// 		}
// 		// 如果key 相等，则找到了目标直接返回
// 		if move.nexts[level] != nil && move.nexts[level].key == key {
// 			return move.nexts[level]
// 		}
// 		// 当前层没找到目标，则层数减 1，继续向下
// 	}
// 	return nil
// }

// // 将 key-val 对加入 skiplist
// func (s *SkipList) Put(key, val int) {
// 	// 假设 kv 对已存在，则直接对值进行更新并返回
// 	if _node := s.search(key); _node != nil {
// 		_node.val = val
// 		return
// 	}
// 	// roll 出新节点的高度
// 	level := s.roll()

// 	// 新节点高度超出跳表最大高度，则需要对高度进行补齐
// 	for len(s.head.nexts)-1 < level {
// 		s.head.nexts = append(s.head.nexts, nil)
// 	}
// 	// 创建出新的节点
// 	newNode := node{
// 		key:   key,
// 		val:   val,
// 		nexts: make([]*node, level+1),
// 	}
// 	// 从头节点的最高层出发
// 	move := s.head
// 	for level := level; level >= 0; level-- {
// 		// 向右遍历，直到右侧节点不存在或者 key 值大于 key
// 		for move.nexts[level] != nil && move.nexts[level].key < key {
// 			move = move.nexts[level]
// 		}
// 		// 调整指针关系，完成新节点的插入
// 		newNode.nexts[level] = move.nexts[level]
// 		move.nexts[level] = &newNode
// 	}
// }

// // 根据 key 从跳表中删除对应的节点
// func (s *SkipList) Del(key int) {
// 	// 如果 kv 对不存在，则无需删除直接返回
// 	if _node := s.search(key); _node == nil {
// 		return
// 	}
// 	// 从头节点的最高层出发
// 	move := s.head
// 	for level := len(s.head.nexts) - 1; level >= 0; level-- {
// 		// 向右遍历，直到右侧节点不存在或者 key 值大于等于 key
// 		for move.nexts[level] != nil && move.nexts[level].key < key {
// 			move = move.nexts[level]
// 		}
// 		// 右侧节点不存在或者 key 值大于 target，则直接跳过
// 		if move.nexts[level] == nil || move.nexts[level].key > key {
// 			continue
// 		}
// 		// 走到此处意味着右侧节点的 key 值必然等于 key，则调整指针引用
// 		move.nexts[level] = move.nexts[level].nexts[level]
// 	}
// 	// 对跳表的最大高度进行更新
// 	var dif int
// 	// 倘若某一层已经不存在数据节点，高度需要递减
// 	for level := len(s.head.nexts) - 1; level > 0 && s.head.nexts[level] == nil; level-- {
// 		dif++
// 	}
// 	s.head.nexts = s.head.nexts[:len(s.head.nexts)-dif]
// }

// // https://mp.weixin.qq.com/s?__biz=MzkxMjQzMjA0OQ==&mid=2247484204&idx=1&sn=54817591aa44359cde9b1b88d386b31b&chksm=c10c4df2f67bc4e416d9f3c23afc989afa56b06afd3c26baeed508a406b885fd1bf361675bcf&scene=21#wechat_redirect

// // roll 骰子，决定一个待插入的新节点在 skiplist 中最高层对应的 index
// func (s *SkipList) roll() int {
// 	var level int
// 	for rand.Int() > 0 {
// 		level++
// 	}
// 	return level
// }
