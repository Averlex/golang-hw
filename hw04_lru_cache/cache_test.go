package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

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
		c := NewCache(3)

		c.Set("key1", 100)
		c.Set("key2", 200)
		c.Set("key3", 300)

		c.Clear()

		notInCacheChecks(t, &c, "key1")
		notInCacheChecks(t, &c, "key2")
		notInCacheChecks(t, &c, "key3")
	})

	t.Run("incorrect capacity", incorrectCapacity)
	t.Run("single element cache", cacheSingleItemSuite)
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
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

func notInCacheChecks(t *testing.T, c *Cache, key Key) {
	t.Helper()

	v, ok := (*c).Get(key)
	require.False(t, ok)
	require.Nil(t, v)
}

func incorrectCapacity(t *testing.T) {
	t.Helper()

	t.Run("zero capacity cache is unusable", func(t *testing.T) {
		c := NewCache(0)
		require.Nil(t, c)
	})

	t.Run("negative capacity cache is unusable", func(t *testing.T) {
		c := NewCache(-1)
		require.Nil(t, c)
	})
}

type CacheTestHelper struct {
	suite.Suite
	cache Cache
}

func (s *CacheTestHelper) isNotInCache(k Key) {
	v, ok := s.cache.Get(k)
	s.False(ok)
	s.Nil(v)
}

func (s *CacheTestHelper) isInCache(k Key, val interface{}) {
	v, ok := s.cache.Get(k)
	s.True(ok)
	s.Equal(val, v)
}

func (s *CacheTestHelper) setExisting(k Key, val interface{}) {
	wasInCache := s.cache.Set(k, val)
	s.True(wasInCache)
}

func (s *CacheTestHelper) setNew(k Key, val interface{}) {
	wasInCache := s.cache.Set(k, val)
	s.False(wasInCache)
}

type SingleItemCacheSuite struct {
	CacheTestHelper
}

func (s *SingleItemCacheSuite) SetupTest() {
	s.cache = NewCache(1)
}

func (s *SingleItemCacheSuite) TestSetToEmpty() {
	s.setNew("key1", 100)
}

func (s *SingleItemCacheSuite) TestSetWithUpdate() {
	s.setNew("key1", 100)
	s.setExisting("key1", 200)
}

func (s *SingleItemCacheSuite) TestSetToFull() {
	s.setNew("key1", 100)
	s.setNew("key2", 200)
}

func (s *SingleItemCacheSuite) TestGetFromEmpty() {
	s.isNotInCache("key1")
}

func (s *SingleItemCacheSuite) TestGetFromFilled() {
	s.cache.Set("key1", 100)
	s.isInCache("key1", 100)
}

func (s *SingleItemCacheSuite) TestGetEvicted() {
	s.cache.Set("key1", 100)
	s.cache.Set("key2", 200)
	s.isNotInCache("key1")
}

func (s *SingleItemCacheSuite) TestGetNonExistent() {
	s.cache.Set("key1", 100)
	s.isNotInCache("key2")
}

func (s *SingleItemCacheSuite) TestClearEmpty() {
	s.cache.Clear()
	s.isNotInCache("key1")
}

func (s *SingleItemCacheSuite) TestClearFilled() {
	s.cache.Set("key1", 100)
	s.cache.Clear()
	s.isNotInCache("key1")
}

func cacheSingleItemSuite(t *testing.T) {
	t.Helper()
	suite.Run(t, new(SingleItemCacheSuite))
}
