package ds

const (
	AL_START_HEAD = 0
	AL_START_TAIL = 1
)

//链表节点
type listNode struct {
	pre   *listNode
	next  *listNode
	value interface{}
}

func (node *listNode) GetValue() (interface{}) {
	return node.value
}

//链表
type list struct {
	head   *listNode
	tail   *listNode
	dup    func(ptr interface{}) interface{}
	free   func(ptr interface{})
	match  func(ptr interface{}, key interface{}) bool
	length int64
}

//创建链表
func ListCreate() *list {
	list := &list{}
	list.head, list.tail = nil, nil
	list.length = 0
	list.dup = nil
	list.free = nil
	list.match = nil

	return list
}

//释放整个链表
func ListRelease(l *list) {
	length := l.length
	for cur := l.head; length > 0; length-- {
		next := cur.next
		if nil != l.free {
			l.free(cur.value)
		}
		cur = next
	}
	l = nil
}

//链表转切片
func (l *list) ToSlice() (s []interface{}) {
	s = make([]interface{}, 0, l.length)
	cur := l.head
	for nil != cur {
		s = append(s, cur.value)
		cur = cur.next
	}
	return s
}

//表头插入节点
func (l *list) AddNodeHead(v interface{}) {
	node := &listNode{value: v}
	if l.length == 0 {
		node.pre, node.next = nil, nil
		l.head, l.tail = node, node
	} else {
		node.pre, node.next = nil, l.head
		l.head.pre = node
		l.head = node
	}
	l.length++
}

//表尾插入节点
func (l *list) AddNodeTail(v interface{}) {
	node := &listNode{value: v}
	if l.length == 0 {
		node.pre, node.next = nil, nil
		l.head, l.tail = node, node
	} else {
		node.pre, node.next = l.tail, nil
		l.tail.next = node
		l.tail = node
	}
	l.length++
}

//创建一个包含值 value 的新节点，并将它插入到 old_node 的之前或之后
func (l *list) InsertNode(oldNode *listNode, value interface{}, after bool) {
	if nil == oldNode {
		return
	}
	node := &listNode{value: value}
	if after {
		node.pre = oldNode
		node.next = oldNode.next
		if l.tail == oldNode {
			l.tail = node
		}
	} else {
		node.pre = oldNode.pre
		node.next = oldNode
		if l.head == oldNode {
			l.head = node
		}
	}
	if nil != node.next {
		node.next.pre = node
	}
	if nil != node.pre {
		node.pre.next = node
	}
	l.length++
}

//删除节点
func (l *list) DeleteNode(node *listNode) {
	if 0 == l.length || nil == node {
		return
	}
	if nil != node.pre {
		node.pre.next = node.next
	} else {
		l.head = node.next
	}
	if nil != node.next {
		node.next.pre = node.pre
	} else {
		l.tail = node.pre
	}
	if nil != l.free {
		l.free(node.value)
	}
	node = nil
	l.length--
}

//通过key查找节点
func (l *list) ListSearchKey(key interface{}) *listNode {
	if nil == l {
		return nil
	}
	//初始化返回节点
	var node *listNode = nil
	//迭代
	iter := ListGetIterator(l, AL_START_HEAD)
	for current:=iter.Next();nil!=current;current=iter.Next() {
		if nil != l.match {
			if l.match(current.GetValue(), key) {
				node = current
				break
			}
		} else {
			if current.GetValue() == key {
				node = current
				break
			}
		}
	}
	ListReleaseIterator(iter)
	return node
}

//返回链表给定索引上的值
func (l *list) Index(index int64) *listNode {
	var node *listNode = nil
	var n int64
	var iter *listIter
	if index < 0 {
		iter = ListGetIterator(l, AL_START_TAIL)
		n = (-index)-1
	} else {
		iter = ListGetIterator(l, AL_START_HEAD)
		n = index
	}
	for current:=iter.Next();nil!=current&&n>=0;current=iter.Next(){
		node = current
		n--
	}
	ListReleaseIterator(iter)
	return node
}

//取出链表的表尾节点，并将它移动到表头，成为新的表头节点
func (l *list) Rotate() {
	if nil == l || l.length <= 1 {
		return
	}

	//取出表尾
	tail := l.tail
	l.tail = tail.pre
	l.tail.next = nil

	//将表尾插到表头
	tail.next = l.head
	tail.pre = nil
	l.head.pre = tail
	l.head = tail
}

//迭代器
type listIter struct {
	next      *listNode
	direction int
}

//获取迭代器
func ListGetIterator(l *list, direction int) *listIter {
	var listIter = &listIter{direction: direction}
	if direction == AL_START_HEAD {
		listIter.next = l.head
	} else {
		listIter.next = l.tail
	}
	return listIter
}

//释放迭代器
func ListReleaseIterator(iter *listIter) {
	iter = nil
}

//将迭代器恢复为头部方向
func (iter *listIter) RewindHead(l *list) {
	iter.next = l.head
	iter.direction = AL_START_HEAD
}

//将迭代器恢复为尾部方向
func (iter *listIter) RewindTail(l *list) {
	iter.next = l.tail
	iter.direction = AL_START_TAIL
}

//获取迭代器下一个节点
func (iter *listIter) Next() *listNode {
	//获取迭代器下一个节点作为当前节点
	current := iter.next

	if nil != current {
		if iter.direction == AL_START_HEAD { //从头方向迭代
			iter.next = current.next
		} else { //从尾方向迭代
			iter.next = current.pre
		}
	}

	return current
}

//复制链表
func ListDup(orig *list) (*list) {
	//源链表为空直接返回
	if nil == orig {
		return nil
	}

	//初始化目的链表
	dest := ListCreate()
	if nil == dest {
		return nil
	}
	dest.dup = orig.dup
	dest.free = orig.free
	dest.match = orig.match

	//通过源链表获取迭代器
	iter := ListGetIterator(orig, AL_START_HEAD)
	for current:=iter.Next();nil!=current;current=iter.Next() {
		var value interface{}
		if nil != orig.dup {
			value = orig.dup(current.GetValue())
		} else {
			value = current.GetValue()
		}
		dest.AddNodeTail(value)
	}

	//释放掉迭代器
	ListReleaseIterator(iter)

	return dest
}