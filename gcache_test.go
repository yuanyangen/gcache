package gcache

import "testing"

func Test_SetAndGet(t *testing.T) {
	key1 := "test_string"
	value1 := "test value"
	_ = Gcache.Set(key1, value1, 0)
	v := Gcache.Get(key1)
	if v != value1 {
		t.Errorf("test get and set string failed\n")
	}



}