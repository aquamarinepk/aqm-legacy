package auth

import (
	"sync"
	"testing"
	"time"
)

func TestNewTTLCache(t *testing.T) {
	cache := NewTTLCache[string, int](5 * time.Minute)

	if cache == nil {
		t.Fatal("NewTTLCache returned nil")
	}
	if cache.defaultTTL != 5*time.Minute {
		t.Errorf("defaultTTL = %v, want 5m", cache.defaultTTL)
	}
	if cache.items == nil {
		t.Error("items map should be initialized")
	}
}

func TestTTLCacheSetAndGet(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	cache.Set("key1", "value1")

	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Get should return true for existing key")
	}
	if val != "value1" {
		t.Errorf("val = %s, want value1", val)
	}
}

func TestTTLCacheGetNonExistent(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	val, ok := cache.Get("nonexistent")
	if ok {
		t.Error("Get should return false for non-existent key")
	}
	if val != "" {
		t.Errorf("val = %s, want empty string", val)
	}
}

func TestTTLCacheGetExpired(t *testing.T) {
	cache := NewTTLCache[string, string](1 * time.Millisecond)

	cache.Set("key1", "value1")
	time.Sleep(5 * time.Millisecond)

	val, ok := cache.Get("key1")
	if ok {
		t.Error("Get should return false for expired key")
	}
	if val != "" {
		t.Errorf("val = %s, want empty string", val)
	}
}

func TestTTLCacheSetWithTTL(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	cache.SetWithTTL("key1", "value1", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Get should return false for expired key set with custom TTL")
	}
}

func TestTTLCacheDelete(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	cache.Set("key1", "value1")
	cache.Delete("key1")

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Get should return false after Delete")
	}
}

func TestTTLCacheDeleteNonExistent(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	// Should not panic
	cache.Delete("nonexistent")
}

func TestTTLCacheClear(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Size = %d, want 0 after Clear", cache.Size())
	}
}

func TestTTLCacheClearExpired(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	cache.SetWithTTL("expired1", "value1", 1*time.Millisecond)
	cache.SetWithTTL("expired2", "value2", 1*time.Millisecond)
	cache.Set("valid", "value3")

	time.Sleep(5 * time.Millisecond)
	cache.ClearExpired()

	if cache.Size() != 1 {
		t.Errorf("Size = %d, want 1 after ClearExpired", cache.Size())
	}

	_, ok := cache.Get("valid")
	if !ok {
		t.Error("valid key should still exist")
	}
}

func TestTTLCacheSize(t *testing.T) {
	cache := NewTTLCache[string, string](5 * time.Minute)

	if cache.Size() != 0 {
		t.Errorf("Size = %d, want 0 for empty cache", cache.Size())
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if cache.Size() != 2 {
		t.Errorf("Size = %d, want 2", cache.Size())
	}
}

func TestTTLCacheConcurrency(t *testing.T) {
	cache := NewTTLCache[int, int](5 * time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Set(n, n*2)
			cache.Get(n)
			cache.Delete(n)
		}(i)
	}
	wg.Wait()
}

func TestNewStringTTLCache(t *testing.T) {
	cache := NewStringTTLCache[int](5 * time.Minute)

	if cache == nil {
		t.Fatal("NewStringTTLCache returned nil")
	}
	if cache.TTLCache == nil {
		t.Error("embedded TTLCache should not be nil")
	}
}

func TestStringTTLCacheDeleteByPrefix(t *testing.T) {
	cache := NewStringTTLCache[string](5 * time.Minute)

	cache.Set("user:123:profile", "profile1")
	cache.Set("user:123:settings", "settings1")
	cache.Set("user:456:profile", "profile2")
	cache.Set("other:data", "data")

	cache.DeleteByPrefix("user:123:")

	if cache.Size() != 2 {
		t.Errorf("Size = %d, want 2 after DeleteByPrefix", cache.Size())
	}

	_, ok := cache.Get("user:123:profile")
	if ok {
		t.Error("user:123:profile should be deleted")
	}

	_, ok = cache.Get("user:456:profile")
	if !ok {
		t.Error("user:456:profile should still exist")
	}

	_, ok = cache.Get("other:data")
	if !ok {
		t.Error("other:data should still exist")
	}
}

func TestStringTTLCacheDeleteByPrefixNoMatch(t *testing.T) {
	cache := NewStringTTLCache[string](5 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	cache.DeleteByPrefix("nonexistent:")

	if cache.Size() != 2 {
		t.Errorf("Size = %d, want 2 (no keys should be deleted)", cache.Size())
	}
}

func TestStringTTLCacheDeleteByPrefixExactMatch(t *testing.T) {
	cache := NewStringTTLCache[string](5 * time.Minute)

	cache.Set("prefix", "value1")
	cache.Set("prefix:extra", "value2")

	cache.DeleteByPrefix("prefix")

	// Only "prefix:extra" should be deleted (key must be longer than prefix)
	if cache.Size() != 1 {
		t.Errorf("Size = %d, want 1", cache.Size())
	}

	_, ok := cache.Get("prefix")
	if !ok {
		t.Error("exact prefix match should not be deleted")
	}
}

func TestTTLCacheWithIntKeys(t *testing.T) {
	cache := NewTTLCache[int, string](5 * time.Minute)

	cache.Set(1, "one")
	cache.Set(2, "two")

	val, ok := cache.Get(1)
	if !ok || val != "one" {
		t.Errorf("Get(1) = %s, %v; want one, true", val, ok)
	}
}

func TestTTLCacheWithStructValues(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	cache := NewTTLCache[string, Person](5 * time.Minute)

	cache.Set("john", Person{Name: "John", Age: 30})

	val, ok := cache.Get("john")
	if !ok {
		t.Error("Get should return true")
	}
	if val.Name != "John" || val.Age != 30 {
		t.Errorf("val = %+v, want {Name:John Age:30}", val)
	}
}
