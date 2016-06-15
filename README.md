# gcache

## features
gcache is a simple cache via golang, these features are finished current:
* 1 get , set, Mget Delete function
* 2 lazy expiration
* 3 LRU-k evict method, you can set the k, queue length


## install
```
go get github.com/yuanyangen/gcache
```

## usage

set :

```
err := gcache.Set("key1", "value1", 0)
```

get
```
// return value type is interface , so you should assert its type
v := gcache.Get("key1")

```
mget
```
v := gcache.MGet([]string{"key1", "key2"})
```

delete
```
error := gcache.Delete("key1")

```