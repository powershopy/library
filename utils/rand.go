package utils

import (
	"github.com/gogf/gf/v2/container/gset"
	"math/rand"
	"time"
)

func RandInt(n int, ignore ...int) int {
	rand.Seed(time.Now().UnixNano())
	s := gset.NewIntSetFrom(ignore)
	for {
		num := rand.Intn(n)
		if !s.Contains(num) {
			return num
		}
	}
}
