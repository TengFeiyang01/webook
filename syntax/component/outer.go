package main

type Inner struct {
}

func (i Inner) DoSomething() {
	println("这是 inner")
}

func (i Inner) SayHello() {
	println("hello", i.Name())
}

func (i Inner) Name() string {
	return "inner"
}

type Outer struct {
	Inner
}

func (o Outer) Name() string {
	return "outer"
}

type OuterV1 struct {
	Inner
}

func (o OuterV1) DoSomething() {
	println("这是 outerv1")
}

type OuterPtr struct {
	*Inner
}

type OOOOuter struct {
	Outer
	OuterV1
}

func UseInner() {
	var o Outer
	o.DoSomething()
	o.Inner.DoSomething()

	var op *OuterPtr
	op.DoSomething()

	o1 := Outer{
		Inner: Inner{},
	}
	op1 := OuterPtr{
		Inner: &Inner{},
	}
	o1.DoSomething()
	op1.DoSomething()
}

func main() {
	var o1 OuterV1
	o1.DoSomething()
	o1.Inner.DoSomething()

	var o Outer
	// 输出什么呢？
	// hello, inner
	// hello, outer
	o.SayHello()
}
