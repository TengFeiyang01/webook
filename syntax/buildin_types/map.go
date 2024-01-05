package main

func Map() {
	m1 := map[string]int{
		"key1": 123,
	}
	m1["hello"] = 345

	// 容量
	m2 := make(map[string]int, 12)
	m2["key2"] = 12

	val, ok := m1["大明"]
	if ok {
		// 有这个键值对
		println(val)
	}

	val = m1["小明"]
	println("小明对应的值:", val)

	delete(m1, "key3")
}
