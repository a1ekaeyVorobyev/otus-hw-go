package hw04_lru_cache //nolint:golint,stylecheck

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"sync"
	"syreclabs.com/go/faker"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c, _ := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c, _ := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		len := 10
		rand.Seed(time.Now().UTC().UnixNano())
		cntRepeat := 3 // rand.Intn(len)
		type People = struct {
			LastName  string
			FirstName string
			DateBorn  time.Time
		}
		c, _ := NewCache(len) //емкость 5 для 10 элементов
		a := make([]Key, len)
		value := make(map[Key]People)
		for i := 0; i < len; i++ {
			email := faker.Internet().Email()
			fio := People{
				LastName:  faker.Name().LastName(),
				FirstName: faker.Name().FirstName(),
				DateBorn:  faker.Date().Birthday(0, 99),
			}
			a[len-i-1] = Key(email)
			value[Key(email)] = fio
			c.Set(Key(email), fio)
		}
		k := c.printCash()
		require.Equal(t, k[0], a[0])
		//get cntRepeat
		for j := 0; j < cntRepeat; j++ {
			// берем из списка случайный элемент
			i := rand.Intn(len - 1)
			fmt.Println("i", i, " key=", a[i])
			val, ok := c.Get(a[i])
			require.Equal(t, ok, true)
			require.Equal(t, val.(People), value[a[i]])
			temp := a[i]
			for l := i; l > 0; l-- {
				a[l] = a[l-1]
			}
			a[0] = temp
			v := c.getFrontElement().(*Item).value.(People)
			k := c.getFrontElement().(*Item).key
			require.Equal(t, a[0], k)
			require.Equal(t, value[a[0]], v)

			// добовляем новый
			email := faker.Internet().Email()
			fio := People{
				LastName:  faker.Name().LastName(),
				FirstName: faker.Name().FirstName(),
				DateBorn:  faker.Date().Birthday(0, 99),
			}
			_, ok = c.Get(Key(email))
			require.Equal(t, ok, false)
			require.Equal(t, c.getBackElement().(*Item).key, a[len-1])
			delete(value, a[len-1])
			for l := len - 1; l > 0; l-- {
				a[l] = a[l-1]
			}
			a[0] = Key(email)
			value[Key(email)] = fio
			ok = c.Set(Key(email), fio)
			v = c.getFrontElement().(*Item).value.(People)
			k = c.getFrontElement().(*Item).key
			require.Equal(t, a[0], k)
			require.Equal(t, value[a[0]], v)
		}
	})
}

func TestCacheMultithreading(t *testing.T) {
	//t.Skip() // Remove if task with asterisk completed

	c, _ := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}

func TestCreateLruCache(t *testing.T) {
	_, err := NewCache(-1)
	require.Equal(t, err, ErrCapacity)
	l, err := NewCache(10)
	require.NoError(t, err, "Error with create Lru Cache")
	err = l.Clear()
	require.Equal(t, err, ErrQueueEmpty)

}
