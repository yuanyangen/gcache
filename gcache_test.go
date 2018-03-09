package gcache

import (
	"testing"
	"fmt"
	"time"
	"strconv"
	"sync"
)

func Test_SetAndGet(t *testing.T) {
	wg := sync.WaitGroup{}
	for i:= 0; i < 100; i++ {
		go func() {
			wg.Add(1)
			key1 := "test_string" + strconv.Itoa(i)
			value1 := "test value"
			_ = Set(key1, value1, 0)
			v := Get(key1)
			if v != value1 {
				t.Errorf("test get and set string failed\n")
			}

			key2 := "test_key2" + strconv.Itoa(i)
			value2 := []string{"addfadsfads","fdasfads"}
			_ = Set(key2, value2, 1)
			tmp2 := Get(key2)
			v2 := tmp2.([]string)
			if v2[0] != value2[0] || v2[1] != value2[1] {
				t.Errorf("test get and set string failed\n")
			}
			time.Sleep(2 * time.Second)
			tmp2 = Get(key2)
			if tmp2 != nil {
				t.Errorf("test get and expire failed\n")
			}
			wg.Done()
		}()
	}
	wg.Wait()
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
