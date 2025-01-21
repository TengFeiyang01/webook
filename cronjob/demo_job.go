package main

import (
	"log"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		log.Println("模拟运行", i)
		time.Sleep(1 * time.Second)
	}
}
