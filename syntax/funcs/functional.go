package main

func Functional4() string {
	println("hello, functional 4")
	return "hello"
}

func Functional5(age int) {

}

var Abc = func() string {
	return "hello"
}

func UseFunctional4() {
	myFunc := Functional4
	myFunc()
	//Abc = func(a int) string {
	//
	//}
	myFunc5 := Functional5
	myFunc5(18)
}

func functional6() {
	// 新定义了一个方法，赋值给了 fn
	fn := func() string {
		return "hello"
	}

	fn()
}

func functional6_1() {
	// 新定义了一个方法，赋值给了 fn
	fn := func() string {
		return "hello"
	}
	fn()
}

func functional8() {
	// 新定义了一个方法，赋值给了 fn
	fn := func() string {
		return "hello"
	}()
	println(fn)
}

// 它的意思是我返回一个，返回 string 的无参数方法
func functional7() func() string {
	return func() string {
		return "hello, world"
	}
}
