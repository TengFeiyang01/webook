package main

import "fmt"

func Byte() {
	var a byte = 'a'
	println(a)
	println(fmt.Sprintf("%c", a))

	var str string = "this is string"
	var bs []byte = []byte(str)
	bs[0] = 'T'
	println(str)
}
