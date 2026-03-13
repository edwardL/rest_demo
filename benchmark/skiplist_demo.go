package main

import (
	"fmt"
	"math/rand"
	"time"
)

const MaxLevel = 32
const P = 0.5

type Node struct {
	value int
	next  []*Node
}

type SkipList struct {
	head  *Node
	level int
}

func NewNode(value, level int) *Node {
	return &Node{
		value: value,
		next:  make([]*Node, level),
	}
}

func NewSkipList() *SkipList {
	return &SkipList{
		head:  NewNode(-1, MaxLevel),
		level: 1,
	}
}

func (sl *SkipList) randomLevel() int {
	level := 1
	for rand.Float64() < P && level < MaxLevel {
		level++
	}
	return level
}

func (sl *SkipList) Search(target int) bool {
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].value < target {
			current = current.next[i]
		}
	}
	current = current.next[0]
	return current != nil && current.value == target
}

func (sl *SkipList) Add(num int) {
	update := make([]*Node, MaxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].value < num {
			current = current.next[i]
		}
		update[i] = current
	}

	level := sl.randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.head
		}
		sl.level = level
	}

	newNode := NewNode(num, level)
	for i := 0; i < level; i++ {
		newNode.next[i] = update[i].next[i]
		update[i].next[i] = newNode
	}
}

func (sl *SkipList) Erase(num int) bool {
	update := make([]*Node, MaxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].value < num {
			current = current.next[i]
		}
		update[i] = current
	}

	current = current.next[0]
	if current == nil || current.value != num {
		return false
	}

	for i := 0; i < sl.level; i++ {
		if update[i].next[i] != current {
			break
		}
		update[i].next[i] = current.next[i]
	}

	for sl.level > 1 && sl.head.next[sl.level-1] == nil {
		sl.level--
	}
	return true
}

func (sl *SkipList) Print() {
	for i := sl.level - 1; i >= 0; i-- {
		current := sl.head.next[i]
		fmt.Printf("Level %d: ", i)
		for current != nil {
			fmt.Printf("%d ", current.value)
			current = current.next[i]
		}
		fmt.Println()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	sl := NewSkipList()

	nums := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}
	for _, num := range nums {
		sl.Add(num)
	}

	fmt.Println("Skip List after insertion:")
	sl.Print()

	fmt.Println("\nSearch for 5:", sl.Search(5))
	fmt.Println("Search for 11:", sl.Search(11))

	fmt.Println("\nErase 5:", sl.Erase(5))
	fmt.Println("Search for 5 after erase:", sl.Search(5))

	fmt.Println("\nSkip List after erasing 5:")
	sl.Print()
}
