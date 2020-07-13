package hw04_lru_cache //nolint:golint,stylecheck
import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type Key string

var _ Cache = (*lruCache)(nil)
var ErrCapacity = errors.New("capacity value must be greater than 0")
var ErrQueueEmpty = errors.New("queue is empty")

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear() error
	printCash() []interface{}
	getFrontElement() interface{}
	getBackElement() interface{}
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*listItem
	mux      *sync.Mutex
}

type Item struct {
	key   Key
	value interface{}
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	if l.capacity == 0 {
		return false
	}
	l.mux.Lock()
	defer l.mux.Unlock()
	if _, ok := l.items[key]; ok { // элемент присутствует в словаре
		l.items[key].data.(*Item).value = value                   //обновили значение
		if err := l.queue.MoveToFront(l.items[key]); err == nil { //переместить элемент в начало очереди
			l.items[key] = l.queue.Front()
			return ok
		}
	} else { //элемента нет в словаре
		l.queue.PushFront(&Item{
			key:   key,
			value: value,
		})
		l.items[key] = l.queue.Front()
		if l.queue.Len() > l.capacity { //размер очереди больше ёмкости кэша
			if err := l.Clear(); err != nil {
				fmt.Printf("%v\r\n", errors.Wrap(err, "can't Set element "+string(key)))
			}
		}
		return ok
	}
	return false
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if _, ok := l.items[key]; ok {
		if err := l.queue.MoveToFront(l.items[key]); err == nil { //переместить элемент в начало очереди
			l.items[key] = l.queue.Front()
			return l.items[key].data.(*Item).value, ok
		}
	}
	return nil, false
}

func (l *lruCache) Clear() error {
	lastItem := l.queue.Back() //последний элемент из очереди
	if lastItem == nil {
		return ErrQueueEmpty
	}
	if item, ok := lastItem.data.(*Item); ok {
		delete(l.items, item.key)                        // удалить его значение из словаря
		if err := l.queue.Remove(lastItem); err != nil { // удалить последний элемент из очереди
			return errors.Wrap(err, "can't clear lruCache")
		}
	}
	return nil
}

func NewCache(capacity int) (Cache, error) {
	if capacity < 0 {
		return nil, ErrCapacity
	}
	cash := &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    map[Key]*listItem{},
		mux:      &sync.Mutex{},
	}
	return cash, nil
}

func (l *lruCache) printCash() []interface{} {
	l.mux.Lock()
	defer l.mux.Unlock()
	v := make([]interface{}, l.queue.Len())
	elem := l.queue.Front()
	for i := 0; i < l.queue.Len(); i++ {
		v[i] = elem.data.(*Item).key
		elem = elem.Next()
	}
	return v
}

func (l *lruCache) getFrontElement() interface{} {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.queue.Front().data.(*Item)
}

func (l *lruCache) getBackElement() interface{} {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.queue.Back().data.(*Item)
}
