package ds

import (
	"fmt"
	"testing"
)

//go test -v adlist.go adlist_test.go --test.run=TestAddNodeHead
func TestAddNodeHead(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeHead(i)
	}
	s := l.ToSlice()
	for i := 0; i < 10; i++ {
		if s[i] != 9-i {
			t.Fatal(s[i], i)
		}
	}
}

//go test -v adlist.go adlist_test.go --test.run=TestAddNodeTail
func TestAddNodeTail(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}
	s := l.ToSlice()
	for i := 0; i < 10; i++ {
		if s[i] != i {
			t.Fatal(s[i], i)
		}
	}
}

//go test -v adlist.go adlist_test.go --test.run=TestInsertNode
func TestInsertNode(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}
	l.InsertNode(l.head.next, 10, true)
	l.InsertNode(l.head.next, 11, false)
	s := l.ToSlice()
	t.Log(s)
}

//go test -v adlist.go adlist_test.go --test.run=TestDeleteNode
func TestDeleteNode(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}
	l.DeleteNode(l.head)
	l.DeleteNode(l.tail)
	s := l.ToSlice()
	t.Log(s)
}

//go test -v adlist.go adlist_test.go --test.run=TestIter
func TestIter(t *testing.T) {
	l := ListCreate()
	for i := 0; i < 10; i++ {
		l.AddNodeTail(i)
	}

	iter := ListGetIterator(l, AL_START_HEAD)
	for current := iter.Next();current!=nil;current=iter.Next() {
		fmt.Println(current.GetValue())
	}
}