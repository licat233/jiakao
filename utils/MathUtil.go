package utils

import (
	"math/rand"
	"sync"
	"time"
)

var (
	//随机数的互斥锁（确保随机数函数不能被并发访问）
	randomMutex sync.Mutex
)

//GetRandomInt 获取随机整数
func GetRandomInt(start, end int) int {
	randomMutex.Lock()
	<-time.After(1 * time.Nanosecond)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := start + r.Intn(end-start+1)
	randomMutex.Unlock()
	return n
}
