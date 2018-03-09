package gcache

import (
	"testing"
	"fmt"
	"strings"
)

func Test_SetAndGet(t *testing.T) {
	/*
	key1 := "test_string"
	value1 := "test value"
	_ = Set(key1, value1, 0)
	v := Get(key1)
	if v != value1 {
		t.Errorf("test get and set string failed\n")
	}

	key2 := "test_key2"
	value2 := []string{"addfadsfads","fdasfads"}
	_ = Set(key2, value2, 1)
	tmp2 := Get(key2)
	v2 := tmp2.([]string)
	if v2[0] != value2[0] || v2[1] != value2[1] {
		t.Errorf("test get and set string failed\n")
	}

	start := time.Now().UnixNano()
	//it will take about 2 second on a 4core+8G +windows7 x64
	for i:=0; i<10000000; i++ {
		tmp2 = Get(key2)
	}
	fmt.Println(time.Now().UnixNano() - start)


	tmp2 = Get(key2)
	if tmp2 != nil {
		t.Errorf("test get and expire failed\n")
	}
	*/
}

func Test_delete(t *testing.T) {
	key := "test1"
	value := "val1"
	Set(key, value, 0)
	v := Get(key)
	if value != v.(string) {
		t.Errorf("error set or get")
	}


	Delete(key)
	v = Get(key)

	if v != nil {
		t.Errorf("error delete")
	}
	fmt.Println(Gcache.lru.queue2.Len())
	fmt.Println(Gcache.lru.queue1.Len())
}
