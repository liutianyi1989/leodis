package ds

import (
	"fmt"
	"strconv"
	"testing"
)

//go test -v adlist.go adlist_test.go --test.run=TestAddNodeHead
func TestAddNodeHead(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeHead(i)
	}
	l.Print(AL_START_HEAD)
}

//go test -v adlist.go adlist_test.go --test.run=TestAddNodeTail
func TestAddNodeTail(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}
	l.Print(AL_START_HEAD)
}

//go test -v adlist.go adlist_test.go --test.run=TestInsertNode
func TestInsertNode(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}
	l.InsertNode(l.head.next, 10, true)
	l.InsertNode(l.head.next, 11, false)
	l.Print(AL_START_HEAD)
}

//go test -v adlist.go adlist_test.go --test.run=TestDeleteNode
func TestDeleteNode(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}
	l.DeleteNode(l.head)
	l.DeleteNode(l.tail)
	l.Print(AL_START_HEAD)
}

//go test -v adlist.go adlist_test.go --test.run=TestIter
func TestIter(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}

	iter := ListGetIterator(l, AL_START_HEAD)
	for current := iter.Next(); current != nil; current = iter.Next() {
		fmt.Println(current.GetValue())
	}
}

//go test -v adlist.go adlist_test.go --test.run=TestSearchKey
func TestSearchKey(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}

	l.match = func(ptr interface{}, key interface{}) bool {
		switch key.(type) {
		case string:
			key, _ = strconv.Atoi(key.(string))
		}

		if ptr == key {
			return true
		}
		return false
	}

	for i := 0; i < 10; i++ {
		node := l.SearchKey(i)
		if nil == node || node.GetValue() != i {
			t.Fail()
		}
	}

	node := l.SearchKey("-1")
	if nil != node {
		t.Fail()
	}

	node = l.SearchKey("11")
	if nil != node {
		t.Fail()
	}
}

//go test -v adlist.go adlist_test.go --test.run=TestIndex
func TestIndex(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}

	node := l.Index(0)
	if node.GetValue() != 0 {
		t.Fail()
	}

	node = l.Index(-1)
	if node.GetValue() != 9 {
		t.Fatal(node.GetValue())
	}
}

//go test -v adlist.go adlist_test.go --test.run=TestRotate
func TestRotate(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}

	l.Rotate()
	l.Print(AL_START_HEAD)

	l.Print(AL_START_TAIL)
}

//go test -v adlist.go adlist_test.go --test.run=TestDup
func TestDup(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}

	l.dup = func(ptr interface{}) interface{} {
		switch ptr.(type) {
		case int:
			ptr = ptr.(int) + 1
		}
		return ptr
	}

	dup := l.Dup()
	dup.Print(AL_START_HEAD)
	fmt.Println("-")
	dup.Print(AL_START_TAIL)
}
