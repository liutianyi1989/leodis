package ds

import "fmt"

const (
	AL_START_HEAD = 0
	AL_START_TAIL = 1
)

//链表节点
type listNode struct {
	pre   *listNode		//前置节点
	next  *listNode		//后置节点
	value interface{}	//节点值
}

func (node *listNode) GetValue() (interface{}) {
	return node.value
}

//持有链表结构
type list struct {
	//指向表头节点
	head   *listNode
	//指向表尾节点
	tail   *listNode
	//记录链表长度
	length int64
	//节点值复制函数
	dup    func(ptr interface{}) interface{}
	//节点值释放函数
	free   func(ptr interface{})
	//节点值对比函数
	match  func(ptr interface{}, key interface{}) bool
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

//释放整个链表，简单置为nil，辅助GC
func ListRelease(l *list) {
	length := l.length
	for cur := l.head; length > 0; length-- {
		next := cur.next
		if nil != l.free {
			l.free(cur.value)
		} else {
			cur.value = nil
		}
		cur = next
	}
	l = nil
}

//打印链表
func (l *list) Print(direction int) {
	//获取迭代器
	iter := ListGetIterator(l, direction)
	for cur:=iter.Next();nil!=cur;cur=iter.Next(){
		fmt.Println(cur.GetValue())
	}
}

//从表头方向插入节点
func (l *list) AddNodeHead(v interface{}) {
	//为节点分配内存
	node := &listNode{value: v}
	if l.length == 0 {//插入空链表
		node.pre, node.next = nil, nil
		l.head, l.tail = node, node
	} else {//插入非空链表
		node.pre, node.next = nil, l.head
		//维护表头
		l.head.pre = node
		l.head = node
	}
	//维护链表长度
	l.length++
}

//从表尾方向插入节点
func (l *list) AddNodeTail(v interface{}) {
	//为节点分配内存
	node := &listNode{value: v}
	if l.length == 0 {//插入空链表
		node.pre, node.next = nil, nil
		l.head, l.tail = node, node
	} else {//插入非空链表
		node.pre, node.next = l.tail, nil
		//维护表尾
		l.tail.next = node
		l.tail = node
	}
	//维护链表长度
	l.length++
}

//创建一个包含值 value 的新节点，并将它插入到 old_node 的之前或之后
func (l *list) InsertNode(oldNode *listNode, value interface{}, after bool) {
	if nil == oldNode {
		return
	}
	//为节点分配内存
	node := &listNode{value: value}
	//执行插入操作，这里只涉及node节点pre和next属性以及表头、表尾属性的维护
	if after { //插入节点到oldNode节点后，即oldNode.next = node
		node.pre = oldNode
		node.next = oldNode.next
		if l.tail == oldNode { //如果oldNode为原表尾，则需要维护表尾属性
			l.tail = node
		}
	} else { //插入节点到oldNode节点前，即oldNode.pre = node
		node.pre = oldNode.pre
		node.next = oldNode
		if l.head == oldNode { //如果oldNode为原表头，则需要维护表头属性
			l.head = node
		}
	}

	//统一处理插入节点前后节点的属性
	//统一维护插入节点的后置节点next属性
	if nil != node.next {
		node.next.pre = node
	}
	//统一维护插入节点的前置节点pre属性
	if nil != node.pre {
		node.pre.next = node
	}
	//维护链表长度
	l.length++
}

//删除节点
func (l *list) DeleteNode(node *listNode) {
	if 0 == l.length || nil == node {
		return
	}
	if nil != node.pre {//节点非表头
		node.pre.next = node.next
	} else {//节点是表头
		l.head = node.next
	}
	if nil != node.next {//节点非表尾
		node.next.pre = node.pre
	} else {//节点是表尾
		l.tail = node.pre
	}
	if nil != l.free {
		l.free(node.value)
	}
	node = nil
	//维护链表长度
	l.length--
}

//通过key查找节点
func (l *list) SearchKey(key interface{}) *listNode {
	if nil == l {
		return nil
	}
	//初始化返回节点
	var node *listNode = nil
	//初始化迭代器
	iter := ListGetIterator(l, AL_START_HEAD)
	//迭代查找
	for current:=iter.Next();nil!=current;current=iter.Next() {
		if nil != l.match {//链表有匹配方法
			if l.match(current.GetValue(), key) {
				node = current
				break
			}
		} else {//链表无匹配方法
			if current.GetValue() == key {
				node = current
				break
			}
		}
	}
	//释放迭代器
	ListReleaseIterator(iter)
	return node
}

//返回链表给定索引上的值
func (l *list) Index(index int64) *listNode {
	var node *listNode = nil
	var n int64
	var iter *listIter
	if index < 0 {//从表尾方向查找
		iter = ListGetIterator(l, AL_START_TAIL)
		n = (-index)-1 //例如：index为-1则n为0
	} else {//从表头方向查找
		iter = ListGetIterator(l, AL_START_HEAD)
		n = index
	}
	//利用迭代器查找
	for current:=iter.Next();nil!=current&&n>=0;current=iter.Next(){
		node = current
		//计数减一，当n为负数时break，则当前node指向头或尾节点
		n--
	}
	if n < 0 { //补偿操作，只要n为负数，则证明out of range，返回nil
		node = nil
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

//复制链表
func (l *list)Dup() (*list) {
	//源链表为空直接返回
	if nil == l {
		return nil
	}

	//为链表分配内存
	dest := ListCreate()
	if nil == dest {
		return nil
	}
	//复制三个函数属性
	dest.dup = l.dup
	dest.free = l.free
	dest.match = l.match

	//通过源链表获取迭代器
	iter := ListGetIterator(l, AL_START_HEAD)
	for current:=iter.Next();nil!=current;current=iter.Next() {
		var value interface{}
		if nil != l.dup {//链表存在复制函数
			value = l.dup(current.GetValue())
		} else {
			value = current.GetValue()
		}
		//从表尾插入
		dest.AddNodeTail(value)
	}

	//释放掉迭代器
	ListReleaseIterator(iter)

	return dest
}

//迭代器
type listIter struct {
	//指向下一个节点，通过Next方法返回
	next      *listNode
	//迭代方向
	direction int
}

//获取迭代器
func ListGetIterator(l *list, direction int) *listIter {
	var listIter = &listIter{direction: direction}
	if direction == AL_START_HEAD {//从表头开始迭代
		listIter.next = l.head
	} else {//从表尾开始迭代
		listIter.next = l.tail
	}
	return listIter
}

//释放迭代器
func ListReleaseIterator(iter *listIter) {
	//简单置为nil，辅助GC
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

//返回迭代器当前所指向的节点
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