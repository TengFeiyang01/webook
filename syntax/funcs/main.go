package main

func main() {
	//name, age := Func10()
	//println(name, age)
	//
	//name1, _ := Func10()
	//println(name1)
	// 使用 := 的前提，就是左边必须有至少一个新变量
	//name1, name2 := Func10()
	//println(name1)
	//println(name2)
	//
	//Func6("Hello", "Ming")

	//UseFunctional4()
	//functional8()

	//fn := Closure("大明")
	// fn 其实已经从 Closure 里面返回了
	// 但是我 fn 还要用到 "大明"
	//println(fn())

	//getAge := Closure1()
	//println("age 是", getAge())
	//println("age 是", getAge())
	//println("age 是", getAge())
	//println("age 是", getAge())

	//getAge = Closure1()
	//println(getAge())
	//println(getAge())
	//println(getAge())
	//println(getAge())
	//println(getAge())

	//Defer()
	//DeferClosure()
	//DeferClosureV1()

	//println(DeferReturn())
	//println(DeferReturnV1())
	DeferClosureLoopV1()
	DeferClosureLoopV2()
	DeferClosureLoopV3()
}
