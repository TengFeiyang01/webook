package main

func Switch(status int) {
	switch status {
	case 0:
		println("初始化")
	case 1:
		println("运行中")
	default:
		println("未知状态")
	}
	println("hello")
}

func SwitchUser(u User) {
	switch u {
	case User{}:
	case User{name: "大明"}:

	}
}

func SwitchBool(age int) {
	switch {
	case age >= 18:
		println("成年人")
	case age > 12:
		println("运行中")
		//default:
		//	println("未知状态")
	}
}

//func SwitchAny(age [3]int) {
//	switch age {
//	case [3]int{18}:
//		println("成年人")
//case age > 12:
//	println("运行中")
//default:
//	println("未知状态")
//}
//}
