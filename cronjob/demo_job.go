package main

import (
	"log"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		log.Println("模拟执行", i)
		time.Sleep(time.Second)
	}
}
