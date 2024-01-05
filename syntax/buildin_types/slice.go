package main

import "fmt"

func Slice() {
	s1 := []int{9, 8, 7}
	fmt.Printf("s1: %v, len=%d, cap=%d \n", s1, len(s1), cap(s1))

	s2 := make([]int, 3, 4)
	fmt.Printf("s2: %v, len=%d, cap=%d \n", s2, len(s2), cap(s2))

	// {0, 0, 0, 0}
	s3 := make([]int, 4)
	fmt.Printf("s3: %v, len=%d, cap=%d \n", s3, len(s3), cap(s3))

	s4 := make([]int, 0, 4)
	s4 = append(s4, 1)
	s4 = append(s4, 2)
	fmt.Printf("s4: %v, len=%d, cap=%d \n", s4, len(s4), cap(s4))
}

func SubSlice() {
	s1 := []int{2, 4, 6, 8, 10}
	s2 := s1[1:3]
	fmt.Printf("s2: %v, len=%d, cap=%d \n", s2, len(s2), cap(s2))

	// 容量就是 start 开始往后，包括原本 s1 的底层数组的元素个数
	s3 := s1[2:]
	fmt.Printf("s3: %v, len=%d, cap=%d \n", s3, len(s3), cap(s3))
	s4 := s1[:3]
	fmt.Printf("s4: %v, len=%d, cap=%d \n", s4, len(s4), cap(s4))
}

func ShareSlice() {
	s1 := []int{1, 2, 3, 4}
	s2 := s1[2:]
	fmt.Printf("s2: %v, len=%d, cap=%d \n", s2, len(s2), cap(s2))

	s2[0] = 99
	fmt.Printf("s2: %v, len=%d, cap=%d \n", s2, len(s2), cap(s2))
	fmt.Printf("s1: %v, len=%d, cap=%d \n", s1, len(s1), cap(s1))

	s2 = append(s2, 199)
	fmt.Printf("s2: %v, len=%d, cap=%d \n", s2, len(s2), cap(s2))
	fmt.Printf("s1: %v, len=%d, cap=%d \n", s1, len(s1), cap(s1))
}
