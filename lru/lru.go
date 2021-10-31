package lru

import (
	"encoding/json"
	"sync"
	"time"
)

func NewLRCache(cap int) *LRUCache {
	return &LRUCache{
		head: nil,
		end:  nil,
		lock: sync.Mutex{},
		hash: map[string]*Node{},
		len:  0,
		cap:  cap,
	}

}

type Node struct {
	key    string
	val    interface{}
	next   *Node
	pre    *Node
	expire int64
}

type LRUCache struct {
	head *Node
	end  *Node
	lock sync.Mutex
	hash map[string]*Node
	len  int
	cap  int
}

func (l *LRUCache) Get(key string) interface{} {

	l.lock.Lock()
	defer l.lock.Unlock()
	if node, ok := l.hash[key]; ok {

		if node.expire != 0 {
			if time.Now().Unix() > node.expire {
				l.clear(node)
				return nil
			}
		}
		l.setFirst(node)
		return node.val
	}

	return nil
}

func (l *LRUCache) clear(node *Node) {

	delete(l.hash, node.key)
	pre := node.pre
	next := node.next
	//只有一个节点
	if pre == nil && next == nil {
		l.len--
		l.head = nil
		l.end = nil
		return
	}
	//头节点过期
	if pre == nil {
		next.pre = nil
		l.head = next
		l.len--
		return
	}
	//尾节点过期
	if next == nil {
		pre.next = nil
		l.end = pre
		l.len--
		return
	}

	pre.next = next
	next.pre = pre
	l.len--

}

func (l *LRUCache) SetEx(key string, val interface{}, expire time.Duration) {

	l.lock.Lock()
	defer l.lock.Unlock()
	//更新
	if exist, ok := l.hash[key]; ok {

		exist.expire = time.Now().Add(expire).Unix()
		exist.val = val
		l.setFirst(exist)
		return
	}
	newNode := &Node{

		key:    key,
		val:    val,
		next:   l.head,
		pre:    nil,
		expire: time.Now().Add(expire).Unix(),
	}

	if l.head != nil {

		l.head.pre = newNode
	}
	l.head = newNode
	//首个节点
	if l.end == nil {

		l.end = newNode
	}
	l.hash[key] = newNode
	l.len++
	if l.len > l.cap {
		l.clear(l.end)
	}

}

func (l *LRUCache) setFirst(node *Node) {
	pre := node.pre
	next := node.next
	//中间位置
	if pre != nil && next != nil {

		pre.next = next
		next.pre = pre
		l.head.pre = node
		node.next = l.head
		node.pre = nil
		l.head = node
	}
	//本身就是第一个
	if pre == nil {
		return
	}

	//最后一个
	if next == nil {

		pre.next = nil
		l.end = pre
		l.head.pre = node
		node.next = l.head
		node.pre = nil
		l.head = node
	}

	return
}

func (l *LRUCache) Out() string {

	l.lock.Lock()
	defer l.lock.Unlock()
	val := map[string]interface{}{}

	for k, v := range l.hash {

		val[k] = v.val
	}
	jsonByte, _ := json.Marshal(val)
	return string(jsonByte)

}
