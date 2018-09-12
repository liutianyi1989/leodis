package ds

import (
	"github.com/liutianyi1989/leodis/util"
)

const (
	//哈希表的初始大小
	DICT_HT_INITIAL_SIZE = 4

	// 操作成功
	DICT_OK = 0
	// 操作失败（或出错）
	DICT_ERR = 1
)

var (
	//指示字典是否启用 rehash 的标识
	dictCanResize = true
	//强制 rehash 的比率
	dictForceResizeRatio = 5
)

//哈希表节点
type dictEntry struct {
	//键
	key interface{}
	//值
	value interface{}
	//指向下个哈希表节点，形成链表
	next *dictEntry
}

//字典类型特定函数
type dictType struct {
	//计算哈希值的函数
	hashFunction func(key interface{}) uint32
	//复制键的函数
	keyDup func(privateData, key interface{}) interface{}
	//复制值的函数
	valDup func(privateData, key interface{}) interface{}
	//对比键的函数
	keyCompare func(privateData, key1, key2 interface{}) bool
	//释放键的函数
	keyDestructor func(privateData, key interface{})
	//释放值的函数
	valDestructor func(privateData, key interface{})
}

//哈希表
type dictht struct {
	//哈希表节点数组
	table []*dictEntry
	//哈希表大小
	size uint32
	//哈希表大小掩码，用于计算哈希值，总是等于size-1
	sizemask uint32
	//哈希表已有节点数量
	used uint32
}

//初始化(重置)哈希表
func (ht *dictht) reset() {
	ht.used = 0
	ht.size = 0
	ht.sizemask = 0
	ht.table = nil
}

//字典
type dict struct {
	//字典类型函数
	dType *dictType
	//私有数据
	privateData interface{}
	//两张哈希表，用于渐进式rehash
	ht [2]*dictht
	//rehash索引,-1代表没有执行rehash
	rehashidx int32
	//目前正在运行的安全迭代器的数量
	iterators uint32
}

//释放节点值
func (d *dict) freeVal(entry *dictEntry) {
	if nil != d.dType && nil != d.dType.valDestructor {
		d.dType.valDestructor(d.privateData, entry.value)
	}
}

//设置给定节点的值
func (d *dict) setVal(entry *dictEntry, val interface{}) {
	if nil != d.dType && nil != d.dType.valDup {
		entry.value = d.dType.valDup(d.privateData, val)
	} else {
		entry.value = val
	}
}

//释放节点键
func (d *dict) freeKey(entry *dictEntry) {
	if nil != d.dType && nil != d.dType.keyDestructor {
		d.dType.keyDestructor(d.privateData, entry.key)
	}
}

//设置给定节点的键
func (d *dict) setKey(entry *dictEntry, key interface{}) {
	if nil != d.dType && nil != d.dType.keyDup {
		entry.key = d.dType.keyDup(d.privateData, key)
	} else {
		entry.key = key
	}
}

//比对两个键
func (d *dict) CompareKeys(key1, key2 interface{}) bool {
	if nil != d.dType && nil != d.dType.keyCompare {
		return d.dType.keyCompare(d.privateData, key1, key2)
	} else {
		return key1 == key2
	}
}

//计算哈希值
func (d *dict) HashKey(key interface{}) uint32 {
	return d.dType.hashFunction(key)
}

//获取节点的键
func (d *dict) GetKey(entry *dictEntry) interface{} {
	return entry.key
}

//获取节点的值
func (d *dict) GetVal(entry *dictEntry) interface{} {
	return entry.value
}

//获取字典总共的桶数量
func (d *dict) Slots() uint64 {
	return uint64(d.ht[0].size) + uint64(d.ht[1].size)
}

//获取字典总共使用的节点数量
func (d *dict) Size() uint64 {
	return uint64(d.ht[0].used) + uint64(d.ht[1].used)
}

//判断字典是否处于rehash进程中
func (d *dict) IsReHashing() bool {
	if d.rehashidx == -1 {
		return true
	}
	return false
}

//初始化字典
func (d *dict) Init(dType *dictType, privateData interface{}) {
	d.ht[0].reset()
	d.ht[1].reset()

	d.dType = dType
	d.privateData = privateData
	d.rehashidx = -1
	d.iterators = 0
}

//创建字典
func DictCreate(dType *dictType, privateData interface{}) *dict {
	d := &dict{}
	d.Init(dType, privateData)
	return d
}

//缩小给定字典
func (d *dict) Resize() int {
	if !dictCanResize || d.IsReHashing() {
		return DICT_ERR
	}

	//计算让比率接近 1：1 所需要的最少节点数量
	minimal := d.ht[0].used
	if minimal < DICT_HT_INITIAL_SIZE {
		minimal = DICT_HT_INITIAL_SIZE
	}

	return d.Expand(minimal)
}

//计算第一个大于等于 size 的 2 的 N 次方，用作哈希表的值
func (d *dict) nextPower(size uint32) uint32 {
	realSize := size
	//不能大于uint32最大值
	if realSize < ^uint32(0) {
		realSize = DICT_HT_INITIAL_SIZE
		for realSize < size {
			realSize *= 2 //2的幂
		}
	} else {
		realSize = ^uint32(0)
	}
	return realSize
}

func (d *dict) Expand(size uint32) int {
	realSize := d.nextPower(size)
	//正在进行rehash
	//或者0号表已经使用的桶数大于realsize
	if d.IsReHashing() || d.ht[0].used > realSize {
		return DICT_ERR
	}

	//新建一张哈希表，用于rehash
	ht := &dictht{}
	ht.size = realSize
	ht.sizemask = ht.size - 1
	ht.used = 0
	//为哈希表分配size个桶的内存
	ht.table = make([]*dictEntry, ht.size)

	if d.ht[0].table == nil { //如果0号表为nil，则表明这是第一次初始化
		d.ht[0] = ht
	} else { //如果0号不为nil,则需要开启渐进式rehash
		//rehash总是将1号表作为新的表
		d.ht[1] = ht
		//设置rehash标识
		d.rehashidx = 0
	}

	return DICT_OK
}

//按步长进行rehash，以桶为单位，每一步都迁移某个桶里的所有哈希节点
//返回1表示rehash未完成
//返回0表示已经完成或者不在rehash进行中
func (d *dict) Rehash(step uint32) int {
	//不在rehash执行中直接返回
	if !d.IsReHashing() {
		return 0
	}

	//按步长进行迁移
	for ; step > 0; step-- {
		//首先判断0号哈希表是否还有已使用的桶，没有的话证明rehash完毕
		if d.ht[0].used == 0 {
			//永远以0号哈希表作为标准表
			d.ht[0] = d.ht[1]
			//初始化1号哈希表，用作下一次rehash
			d.ht[1].reset()
			//设置rehash表示为-1表示未执行rehash
			d.rehashidx = -1
			return 1
		}

		//如果0号哈希表的rehash索引大于表长度，说明有问题
		if d.ht[0].size < uint32(d.rehashidx) {
			d.rehashidx = 0
			return 0
		}

		//rehash没有完毕，则按rehashidx的值遍历0号哈希表中的桶
		//找到下一个不为空的桶
		for ; d.ht[0].table[d.rehashidx] == nil; d.rehashidx++ {
		}
		//获取桶里面的第一个节点
		for de := d.ht[0].table[d.rehashidx]; de != nil; {
			//需要把0号哈希表de的next节点临时保存下来
			oldNextDe := de.next
			//计算hash索引
			index := d.HashKey(de.key) & d.ht[1].sizemask
			//插入到1号哈希表的index号桶的表头
			de.next = d.ht[1].table[index]
			d.ht[1].table[index] = de
			//维护两张表的used字段
			d.ht[0].used--
			d.ht[1].used++
			//de设置为0号表的next节点
			de = oldNextDe
		}

		//将0号哈希表已迁移的桶置空
		d.ht[0].table[d.rehashidx] = nil
		//更新rehashidx，指向下一个桶
		d.rehashidx++
	}

	//走到这里证明还没有rehash完成
	return 0
}

//在超时时间范围内尽可能多的进行rehash
func (d *dict) RehashMilliseconds(timeoutms int64) uint32 {
	//获取毫秒为单位的时间戳
	startTime := util.TimeInMilliseconds()
	var rehashTimes uint32 = 0
	for d.Rehash(100) != 0 { //每次100步
		rehashTimes += 100
		if util.TimeInMilliseconds()-startTime > timeoutms { //超时即退出
			break
		}
	}
	return rehashTimes
}

//单步执行rehash
//不能再字典有安全迭代器的情况下执行，因为两种操作会导致冲突，以致字典混乱
func (d *dict) rehashStep() {
	if d.iterators == 0 {
		d.Rehash(1)
	}
}

/*
 * 尝试将给定键值对添加到字典中
 * 最坏 T = O(N) ，平均 O(1)
 */
func (d *dict) Add(key, value interface{}) int {
	//判断key是否已经存在
	de := d.AddRaw(key)

	//如果存在则返回DICT_ERR
	if nil == de {
		return DICT_ERR
	}

	//设置值
	d.setVal(de, value)

	return DICT_OK
}

/*
 * 尝试将键插入到字典中
 * 如果键已经在字典存在，那么返回nil
 * 如果键不存在，那么程序创建新的哈希节点，
 * 将节点和键关联，并插入到字典，然后返回节点本身。
 * T = O(N)
 */
func (d *dict) AddRaw(key interface{}) *dictEntry {
	//如果字典正在rehash进程中
	//则执行单步rehash
	if d.IsReHashing() {
		d.rehashStep()
	}

	//键存在则直接返回
	if d.keyIndex(key) < 0 {
		return nil
	}

	//为该节点分配空间
	de := &dictEntry{}
	d.setKey(de, key)

	//如果在rehash进程中，则插入1号表，否则插入0号表
	var ht *dictht
	if d.IsReHashing() {
		ht = d.ht[1]
	} else {
		ht = d.ht[0]
	}
	//计算桶的索引
	idx := d.HashKey(key) & ht.sizemask
	//插入到该桶的头部
	de.next = ht.table[idx]
	ht.table[idx] = de
	//更新节点数
	ht.used++

	return de
}

/*
 * 对addRaw的封装
 * 如果key存在则返回，否则执行addRaw
 */
func (d *dict) ReplaceRaw(key interface{}) *dictEntry {
	de := d.Find(key)
	if nil == de {
		return d.AddRaw(key)
	}
	return de
}

/*
 * 用新值替换key的原值
 * 如果key是新的，则返回1
 * 如果key是旧的，则返回0
 */
func (d *dict) Replace(key, value interface{}) int {
	//先尝试添加一个节点进去
	if d.Add(key, value) == DICT_OK {
		return 1
	}

	//如果节点存在，找到这个节点
	de := d.Find(key)
	//新值覆盖原值
	d.setVal(de, value)
	return 0
}

/*
 * 通过key查找entry
 */
func (d *dict) Find(key interface{}) *dictEntry {
	//未初始化的字典直接返回
	if d.ht[0].size == 0 {
		return nil
	}

	//如果允许的话，执行单步rehash
	if d.IsReHashing() {
		d.rehashStep()
	}

	//计算key对应的哈希值
	h := d.HashKey(key)

	//返回的entry
	var de *dictEntry = nil

	//在哈希表中进行查找
	for tableNum := 0; tableNum <= 1; tableNum++ {
		//计算索引值
		idx := h & d.ht[tableNum].sizemask
		//在idx这个桶中查找
		for current := d.ht[tableNum].table[idx]; nil != current; current = current.next {
			if d.CompareKeys(current.key, key) {
				de = current
				break
			}
		}

		//如果没有在rehash进程中，则无需查找1号哈希表
		if !d.IsReHashing() {
			break
		}
	}

	return de
}

/*
 * 通过key查找value
 */
func (d *dict) FetchValue(key interface{}) interface{} {
	de := d.Find(key)
	if nil == de {
		return nil
	}
	return d.GetVal(de)
}

/*
 * 根据key删除节点
 */
func (d *dict) genericDelete(key interface{}, nofree bool) int {
	//如果字典还没有初始化，直接返回
	if d.ht[0].size == 0 {
		return DICT_ERR
	}

	//如果条件允许，执行单步rehash
	if d.IsReHashing() {
		d.rehashStep()
	}

	//通过key计算哈希值
	h := d.HashKey(key)

	//在表中进行查找
	for tableNum := 0; tableNum <= 1; tableNum++ {
		//计算桶的idx
		idx := h & d.ht[tableNum].sizemask
		//在该桶中进行查找
		var pre *dictEntry = nil
		var current = d.ht[tableNum].table[idx]
		for nil != current {
			if d.CompareKeys(current.key, key) { //found
				if pre == nil { //删除的是该桶的头节点
					d.ht[tableNum].table[idx] = current.next
				} else {
					pre.next = current.next
				}

				//释放删除节点的内存空间
				if !nofree {
					d.freeKey(current)
					d.freeVal(current)
				}
				current = nil

				//更新现有节点数
				d.ht[tableNum].used--

				return DICT_OK
			}
			//节点前进
			pre = current
			current = current.next
		}

		//只有在rehash中，才有必要查1号哈希表
		if !d.IsReHashing() {
			break
		}
	}

	//没找到该节点
	return DICT_ERR
}

/*
 * 返回可以将 key 插入到哈希表的索引位置
 * 如果 key 已经存在于哈希表，那么返回 -1
 * T = O(N)
 */
func (d *dict) keyIndex(key interface{}) uint32 {
	//判断字典是否需要扩容
	d.expandIfNeeded()

	var idx uint32
	//计算要查找key的哈希值
	h := d.HashKey(key)
	//有可能需要查找两张表
	for tableNum := 0; tableNum <= 1; tableNum++ {
		//计算该哈希表的桶索引
		idx = h & d.ht[tableNum].sizemask
		//遍历查找桶中结点
		for de := d.ht[tableNum].table[idx]; nil != de; de = de.next {
			if d.CompareKeys(de.key, key) {
				return -1
			}
		}
		//走到这里说明0号哈希表已经找完了
		//如果没有在rehash进程中，也就没必要再找1号哈希表了
		if !d.IsReHashing() {
			break
		}
	}

	return idx
}

/*
 * 判断字典是否需要扩容或者初始化0号哈希表
 * T = O(N)
 */
func (d *dict) expandIfNeeded() int {
	//如果字典正在rehash中，直接返回
	//T=O(1)
	if d.IsReHashing() {
		return DICT_OK
	}

	//如果0号哈希表长度为0，则直接初始化
	//T=O(1)
	if d.ht[0].size == 0 {
		return d.Expand(DICT_HT_INITIAL_SIZE)
	}

	//是否需要扩容标识
	var ifNeedExpand = false
	if dictCanResize && d.ht[0].used >= d.ht[0].size { //第一种情况：打开扩容开关，且哈希因子大于1
		ifNeedExpand = true
	} else if d.ht[0].used/d.ht[0].size >= uint32(dictForceResizeRatio) { //第二种情况：哈希因子大于5，强制rehash
		ifNeedExpand = true
	}
	//需要扩容
	if ifNeedExpand {
		//0号表使用节点数两倍扩容
		//T=O(N)
		return d.Expand(d.ht[0].used * 2)
	}

	return DICT_OK
}

//打开resize开关
func DictEnableResize() {
	dictCanResize = true
}

//关闭resize开关
func DictDisableResize() {
	dictCanResize = false
}
