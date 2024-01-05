package main

// 递归使用不当，就有可能 stack overflow
func Recursive(n int) {
	if n > 10 {
		return
	}
	Recursive(n + 1)
}

func A() {
	B()
}

func B() {
	C()
}

func C() {
	A()
}
