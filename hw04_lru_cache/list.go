package hw04_lru_cache //nolint:golint,stylecheck
import (
	"github.com/pkg/errors"
)

var _ List = (*list)(nil)
var ErrItemRemove = errors.New("This element has already been deleted")
var ErrItemList = errors.New("This item does not belong to this list")

type List interface {
	Len() int
	Front() *listItem
	Back() *listItem
	PushFront(v interface{})
	PushBack(v interface{})
	Remove(i *listItem) error
	MoveToFront(i *listItem) error
	GetListItem(v interface{}) []*listItem
}

type list struct {
	head *listItem
	tail *listItem
	size int
}

type listItem struct {
	data interface{}
	next *listItem
	prev *listItem
	list *list
}

func (i *listItem) Prev() *listItem {
	return i.prev
}

func (i *listItem) Next() *listItem {
	return i.next
}

func (i *listItem) GetList() *list {
	return i.list
}

func (i *listItem) Value() interface{} {
	return i.data
}

func (l *list) Len() int {
	return l.size
}

func (l *list) PushBack(data interface{}) {
	temp := l.Back()
	item := new(listItem)
	item.data = data
	item.prev = temp
	l.tail = item
	if temp != nil {
		temp.next = item
	}
	if l.Front() == nil {
		l.head = item
	}
	item.list = l
	l.size++
}

func (l *list) PushFront(data interface{}) {
	temp := l.Front()
	item := new(listItem) //:= new(Item)
	item.data = data
	item.prev = nil
	item.next = temp
	if temp != nil {
		temp.prev = item
	}
	l.head = item
	if l.Back() == nil {
		l.tail = item
	}
	item.list = l
	l.size++
}

func (l *list) Front() *listItem {
	return l.head
}

func (l *list) Back() *listItem {
	return l.tail
}

func (l *list) Remove(remoteItem *listItem) error {
	if remoteItem.list == nil {
		return ErrItemRemove
	}
	if l != remoteItem.list {
		return ErrItemList
	}
	remoteItem.list = nil
	prev := remoteItem.prev
	next := remoteItem.next
	if prev != nil {
		prev.next = next
	} else {
		l.head = next
	}
	if next != nil {
		next.prev = prev
	} else {
		l.tail = prev
	}
	l.size--
	return nil
}

func (l *list) MoveToFront(item *listItem) error {

	if err := l.Remove(item); err != nil {
		return err
	}
	l.PushFront(item.data)
	return nil
}

func (l *list) GetListItem(data interface{}) []*listItem {
	temp := make([]*listItem, 0, l.size)
	for e := l.head; e != nil; e = e.Next() {
		if e.data == data {
			temp = append(temp, e)
		}

	}
	return temp
}

func NewList() List {
	return &list{
		tail: nil,
		head: nil,
		size: 0,
	}
}
