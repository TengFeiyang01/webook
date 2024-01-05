package main

const External = "包外"
const internal = "包内"
const (
	a = 123
)

const (
	Init = iota
	Running
	Paused
	Stop

	StatusE = 100
	StatusF
)
const (
	DayA = iota
	// 1 左移
	DayB
	// 2 左移
	DayC
	// 3 左移
	DayD
	DayE
)

func main() {
	const a = 123
}
