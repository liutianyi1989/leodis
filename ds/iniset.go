package ds

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	INTSET_ENC_INT16 = 1
	INTSET_ENC_INT32 = 2
	INTSET_ENC_INT64 = 3

	//int16最大值
	INT16_MAX = int16(^uint16(0) >> 1)
	//int16最小值
	INT16_MIN = ^int16(^uint16(0) >> 1)

	//int32最大值
	INT32_MAX = int32(^uint32(0) >> 1)
	//int32最小值
	INT32_MIN = ^int32(^uint32(0) >> 1)

	//int64最大值
	INT64_MAX = int64(^uint64(0) >> 1)
	//int32最小值
	INT64_MIN = ^int64(^uint64(0) >> 1)
)

type IntSet struct {
	//编码类型
	encoding uint8
	//元素数量
	length uint32
	//元素数组
	contents []byte
}

//辅助方法，获取某个元素的编码类型
func getValueEncoding(v interface{}) uint8 {
	var x int64
	switch v.(type) {
	case int16:
		x = int64(v.(int16))
	case int32:
		x = int64(v.(int32))
	case int64:
		x = v.(int64)
	default:
		panic("不支持的整数集合编码类型")
	}

	var encoding uint8
	if int64(INT16_MIN) <= x && x <= int64(INT16_MAX) {
		encoding = INTSET_ENC_INT16
	} else if int64(INT32_MIN) <= x && x <= int64(INT32_MAX) {
		encoding = INTSET_ENC_INT32
	} else {
		encoding = INTSET_ENC_INT64
	}
	return encoding
}

//整数集合支持的编码转换为字节数组
func IntToBytes(item interface{}) []byte {
	var buffer bytes.Buffer
	switch item.(type) {
	case int16:
		binary.Write(&buffer, binary.LittleEndian, item.(int16))
	case int32:
		binary.Write(&buffer, binary.LittleEndian, item.(int32))
	case int64:
		binary.Write(&buffer, binary.LittleEndian, item.(int64))
	default:
		panic("不支持的整数集合编码类型")
	}
	return buffer.Bytes()
}

//字节数组转换为整数集合支持的编码
func BytesToInt(b []byte, encoding uint8) interface{} {
	bin_buf := bytes.NewBuffer(b)
	var i16 int16
	var i32 int32
	var i64 int64
	var ret interface{}
	switch encoding {
	case INTSET_ENC_INT16:
		binary.Read(bin_buf, binary.LittleEndian, &i16)
		ret = i16
	case INTSET_ENC_INT32:
		binary.Read(bin_buf, binary.LittleEndian, &i32)
		ret = i32
	case INTSET_ENC_INT64:
		binary.Read(bin_buf, binary.LittleEndian, &i64)
		ret = i64
	default:
		panic("不支持的整数集合编码类型")
	}
	return ret
}

//为整数集合分配内存
func NewIntSet() *IntSet {
	intSet := &IntSet{}
	//默认16位
	intSet.encoding = INTSET_ENC_INT16
	intSet.length = 0
	return intSet
}

//获取指定位置上的元素
//T=O（1）
func (this *IntSet) getEncoded(pos uint32, encoding uint8) int64 {
	var b []byte
	var ret int64
	switch encoding {
	case INTSET_ENC_INT16:
		//16位占两个字节
		b = this.contents[2*pos : 2*(pos+1)]
		ret = int64(BytesToInt(b, encoding).(int16))
	case INTSET_ENC_INT32:
		//32位占四个字节
		b = this.contents[4*pos : 4*(pos+1)]
		ret = int64(BytesToInt(b, encoding).(int32))
	case INTSET_ENC_INT64:
		//64位占八个字节
		b = this.contents[8*pos : 8*(pos+1)]
		ret = BytesToInt(b, encoding).(int64)
	default:
		panic("不支持的整数集合编码类型")
	}

	return ret
}

func (this *IntSet) Get(pos uint32) int64 {
	return this.getEncoded(pos, this.encoding)
}

//调整整数集合的内存空间大小
func (this *IntSet) resize(length uint32) {
	var size uint32
	switch this.encoding {
	case INTSET_ENC_INT16:
		//16位占两个字节
		size = 2 * length
	case INTSET_ENC_INT32:
		//32位占四个字节
		size = 4 * length
	case INTSET_ENC_INT64:
		//64位占八个字节
		size = 8 * length
	default:
		panic("不支持的整数集合编码类型")
	}

	//将原有字节数组拷贝到新的数组中，copy函数会取dst和src中的最小长度
	//达到保存原数据的目的
	old := this.contents
	this.contents = make([]byte, size, size)
	copy(this.contents, old)

	////更新元素数量
	//this.length = length
}

//在集合中插入元素
//返回是否插入成功
func (this *IntSet) Add(value int64) (success bool) {
	//默认插入成功
	success = true

	if getValueEncoding(value) > this.encoding { //如果编码类型大于当前集合编码类型，则需要升级

	} else {
		//先在集合中找这个值，如果存在直接返回
		found, _ := this.Search(value)
		if found {
			return false
		}

		//为插入的值分配内存
		this.resize(this.length + 1)
	}

	return success
}

//在字节数组中移动元素
//from为起始索引，to为目标索引
//from<to 扩大
//from>to 缩小
func (this *IntSet) Move(fromPos uint32, toPos uint32) {
	if fromPos == toPos {
		return
	}

	var bLength uint32
	switch this.encoding {
	case INTSET_ENC_INT16:
		//16位占两个字节
		bLength = 2
	case INTSET_ENC_INT32:
		//32位占四个字节
		bLength = 4
	case INTSET_ENC_INT64:
		//64位占八个字节
		bLength = 8
	default:
		panic("不支持的整数集合编码类型")
	}

	if fromPos < toPos {
		start := fromPos * bLength
		end := toPos * bLength
		offset := end - start
		fmt.Println(start, end, offset)
		for i := end - 1; i >= start; i-- {
			this.contents[i+offset] = this.contents[i]
		}
	} else {
		start := fromPos * bLength
		end := toPos * bLength
		offset := (end - start) * bLength
		for i := start; i < this.length; i++ {
			this.contents[i-offset] = this.contents[i]
		}
	}
}

//在集合中查找value
//如果存在则返回value所在索引位置
func (this *IntSet) Search(value int64) (found bool, pos *uint32) {
	//如果集合中没有元素，则直接返回结果
	if this.length == 0 {
		return false, nil
	}

	//因为整数集合是按从小到大排列
	//所以先做边界检查，如果不在范围内，则直接返回未找到
	minV := this.Get(0)
	maxV := this.Get(this.length - 1)
	if value < minV || value > maxV {
		return false, nil
	}

	//二分查找
	//T=O（lgN）
	min := uint32(0)
	max := this.length - 1
	for max >= min {
		mid := (max + min) / 2
		midV := this.Get(mid)
		if value < midV {
			max = mid - 1
		} else if value > midV {
			min = mid + 1
		} else {
			*pos = mid
			return true, pos
		}
	}

	return false, nil
}
