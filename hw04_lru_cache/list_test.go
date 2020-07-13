package hw04_lru_cache //nolint:golint,stylecheck
import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, l.Len(), 0)
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, l.Len(), 3)

		middle := l.Front().Next() // 20
		l.Remove(middle)           // [10, 30]
		require.Equal(t, l.Len(), 2)

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, l.Len(), 7)
		require.Equal(t, 80, l.Front().Value())
		require.Equal(t, 70, l.Back().Value())

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next() {
			elems = append(elems, i.Value().(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func Test_CheckCreate(t *testing.T) {

	checkValue := []string{"dd", "4", "5", "true"}
	s := NewList()
	s.PushFront("dd")
	s.PushFront("4")
	s.PushFront(5)
	s.PushFront(true)
	if s.Len() != 4 {
		t.Error("Не верный размер ", s.Len())
	}
	count := 0

	for e := s.Back(); e != nil; e = e.Prev() {
		st := fmt.Sprintf("%v", e.Value())
		if st != checkValue[count] {

			t.Error(e.Value(), " Не cooтветствует ", checkValue[count])
		}
		count++
	}
	if s.Front().Value() != true {
		t.Error(s.Front().Value(), " не cooтветствует true")
	}
	if s.Back().Value() != "dd" {
		t.Error(s.Back().Value(), " не cooтветствует dd")
	}
}

func Test_Add(t *testing.T) {

	checkValue := []interface{}{true, 5, "4", "dd"}
	s := NewList()
	s.PushFront("dd")
	s.PushFront("4")
	s.PushFront(5)
	s.PushFront(true)
	count := 0
	for e := s.Front(); e != nil; e = e.Next() {
		//st := fmt.Sprintf("%v", e.Value())
		require.Equal(t, checkValue[count], e.Value())
		count++
	}
	checkValue = []interface{}{"sds", true, 5, "4", "dd", 15}
	s.PushFront("sds")
	s.PushBack(15)
	count = 0
	for e := s.Front(); e != nil; e = e.Next() {
		require.Equal(t, checkValue[count], e.Value())
		count++
	}
}

func Test_Remove(t *testing.T) {
	s := NewList()
	s.PushBack("dd")
	s.PushBack("4")
	s.PushBack(true)
	s.PushBack(5)
	s.PushBack(true)
	a := s.GetListItem(true)
	s.MoveToFront(a[0])
	e := s.Front()
	require.Equal(t, true, e.Value())
	lenList := s.Len() - 1
	assert.Nil(t, s.Remove(e))
	err := s.Remove(e)
	require.Equal(t, lenList, s.Len())
	require.Equal(t, err, ErrItemRemove)
	b := NewList()
	b.PushBack("dd")
	err = s.Remove(b.Front())
	require.Equal(t, err, ErrItemList)
}
