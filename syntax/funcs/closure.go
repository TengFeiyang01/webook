package main

import "fmt"

func Closure(name string) func() string {
	return func() string {
		return "hello, " + name
	}
}

func Closure1() func() int {
	var age = 0
	fmt.Printf("out: %p", &age)
	return func() int {
		fmt.Printf("before %p ", &age)
		age++
		fmt.Printf("after %p ", &age)
		return age
	}
}

func DeferClosureLoopV1() {
	for i := 0; i < 10; i++ {
		defer func() {
			fmt.Printf("i 的地址是 %p，值是 %d \n", &i, i)
		}()
	}
}

func DeferClosureLoopV2() {
	for i := 0; i < 10; i++ {
		defer func(val int) {
			fmt.Printf("val 的地址是 %p，值是 %d \n", &val, val)
		}(i)
	}
}

func DeferClosureLoopV3() {
	for i := 0; i < 10; i++ {
		j := i
		defer func() {
			fmt.Printf("j 的地址是 %p，值是 %d \n", &j, j)
		}()
	}
}
