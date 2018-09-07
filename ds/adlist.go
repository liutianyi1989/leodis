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

//链表
type list struct {
	head   *listNode
	tail   *listNode
	dup    func(ptr interface{})
	free   func(ptr interface{})
	match  func(ptr interface{}, key interface{}) bool
	length int64
}

//迭代器
type listIter struct {
	next      *listNode
	direction int
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

//func ListGetIterator(l *list, direction int) *listIter {
//
//}
