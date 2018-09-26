package ds

import "math/rand"

const (
	//跳表最大层数
	ZSKIPLIST_MAXLEVEL = 32
)

//跳表
type SkipList struct {
	//头节点
	head *SkipListNode
	//尾节点
	tail *SkipListNode
	//节点数
	length uint64
	//节点最大层数
	level int
}

//节点
type SkipListNode struct {
	//层
	level []*SkipListLevel
	//后退指针
	backward *SkipListNode
	//分值
	score int64
	//数据
	obj interface{}
}

//层
type SkipListLevel struct {
	//前进指针
	forward *SkipListNode
	//跨度
	span uint64
}

//创建跳表
func ZslCreate() *SkipList {
	//为跳表分配内存
	skiplist := &SkipList{}
	skiplist.tail = nil
	skiplist.level = 1
	skiplist.length = 0
	//设置头结点
	skiplist.head = &SkipListNode{
		//头结点包含最大层数的节点数组，持有各层的起始节点
		level:    make([]*SkipListLevel, 0, ZSKIPLIST_MAXLEVEL),
		backward: nil,
		score:    0,
		obj:      nil,
	}
	return skiplist
}

//释放节点内存
func ZslFreeNode(n *SkipListNode) {
	n = nil
}

//释放跳表
func ZslFree(zsl *SkipList) {
	//从第一层遍历，则跳表退化为双端链表
	cur := zsl.head.level[0].forward
	for nil != cur {
		next := cur.level[0].forward
		ZslFreeNode(cur)
		cur = next
	}
	zsl = nil
}

//插入节点
func (this *SkipList) InsertNode(v interface{}, score int64) {
	//保存每一层插入节点的前驱
	update := make([]*SkipListNode, 0, ZSKIPLIST_MAXLEVEL)
	//保存每一层的rank值
	rank := make([]uint64, 0, ZSKIPLIST_MAXLEVEL)

	//在没有插入节点之前，先初始化update和rank
	x := this.head
	for i := this.level - 1; i >= 0; i-- {
		if i == this.level-1 { //如果是顶层，初始化rank[i]为0
			rank[i] = 0
		} else { //否则逐层累加
			rank[i] = rank[i+1]
		}

		//找到插入节点的位置，第一个分数或者值字典序大于插入节点的（这里先只简单对比分数，相等的按时间先后依次插入）
		for nil != x.level[i].forward && x.level[i].forward.score < score {
			//维护rank值，向前进相当于增加了span个长度
			rank[i] = x.level[i].span
			//按层前进
			x = x.level[i].forward
		}
		//更新update，维护插入节点在每层的前驱
		update[i] = x
	}

	//随机获取一个插入节点的层数(越高的层出现的概率越小)
	level := this.randomLevel()

	//如果新节点的level比插入前最大level要大，需要(this.level,level]区间内的层初始化
	//由下至上
	if this.level < level {
		for i := this.level - 1; i < level; i++ {
			//维护头节点各层
			this.head.level[i] = &SkipListLevel{
				forward: nil,         //前进指针为空
				span:    this.length, //跨度为节点长度
			}
			//初始化update,指向头节点
			update[i] = this.head
			//初始化rank值
			rank[i] = 0
		}
		//更新表最大level属性
		this.level = level
	}

	//为新节点分配内存空间
	newNode := &SkipListNode{
		level:    make([]*SkipListLevel, 0, level),
		backward: nil,
		score:    score,
		obj:      v,
	}

	//由于之前已经在update中维护了新节点在每层中的前驱
	//所有由下至上依次维护每层链表
	for i := 0; i < level; i++ {
		//插入节点
		newNode.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = newNode
		//计算新节点在该层的span
		newNode.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		//计算新节点在该层前驱的span
		update[i].level[i].span = (update[i].level[i].span + 1) - newNode.level[i].span
	}

	//如果有未接触新节点的层，跨度也需要增加1个长度
	for i := level; i < this.level-1; i++ {
		update[i].level[i].span++
	}

	//维护后退指针
	if update[0] == this.head { //插入到第一个
		newNode.backward = nil
	} else {
		newNode.backward = update[0]
	}
	//维护插入后的后继节点的后退指针
	if nil != newNode.level[0].forward {
		newNode.level[0].forward.backward = newNode
	} else { //插入到最后一个
		this.tail = newNode
	}

	//维护跳表长度属性
	this.length++

	return
}

func (this *SkipList) randomLevel() int {
	return rand.Intn(ZSKIPLIST_MAXLEVEL) + 1
}

//内部删除节点函数
//update数组中已经存放了x在每层中的前驱
func (this *SkipList) deleteNode(x *SkipListNode, update []*SkipListNode) {
	//删除节点
	for i := 0; i < this.level; i++ { //与删除节点有接触的层
		if update[i].level[i].forward == x {
			update[i].level[i].forward = x.level[i].forward
			update[i].level[i].span = update[i].level[i].span + x.level[i].span - 1
		} else { //与删除节点没有接触的层
			update[i].level[i].span -= 1
		}
	}

	//维护后退指针
	if nil != x.level[0].forward {
		x.level[0].forward.backward = x.backward
	}
	//如果删除的是尾节点
	if this.tail == x {
		this.tail = x.backward
	}

	//维护跳表最大层数
	for this.level > 1 && this.head.level[this.level-1].forward == nil {
		this.level--
	}

	//维护跳表长度
	this.length--
}

//从链表中删除某个节点
//0:未找到节点
//1:删除成功
func (this *SkipList) Delete(v interface{}, score int64) int {
	//分配内存，维护删除节点在各层的前驱
	update := make([]*SkipListNode, 0, this.level)
	//遍历跳表
	x := this.head
	for i := this.level - 1; i >= 0; i-- {
		for nil != x.level[i].forward &&
			(x.level[i].forward.score < score || (x.level[i].forward.score == score && v != x.level[i].forward.obj)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	//检查是否是要删除的节点
	if x.level[0].forward.score == score && x.level[0].forward.obj == v {
		this.deleteNode(x.level[0].forward, update)
		ZslFreeNode(x.level[0].forward)
		return 1
	}

	return 0
}

//查找指定节点在跳表中的排行值
//由于跨度计算包含表头，所以排行值初始值为1
//返回0代表没有找到节点
func (this *SkipList) GetRank(v interface{}, score int64) uint64 {
	var rank uint64

	//从表头遍历
	x := this.head
	//从顶层由上至下
	for i := this.level - 1; i >= 0; i-- {
		for nil != x.level[i].forward &&
			(score > x.level[i].forward.score || (score == x.level[i].forward.score && v != x.level[i].forward.obj)) {
			//将跨度值累加到rank值
			rank += x.level[i].span
			//在该层前进
			x = x.level[i].forward
		}
	}

	//确认该节点是要定位的节点
	x = x.level[0].forward
	if v == x.obj && score == x.score {
		return rank
	}

	//没找到指定节点
	return 0
}
