package main

import (
	"encoding/json"
)

func Sum[T Number](vals ...T) T {
	var res T
	for _, val := range vals {
		res = res + val
	}
	return res
}

func SumV1[T Number](vals ...T) T {
	var t T
	return t
}

type Number interface {
	~int | int64 | float64 | float32 | int32 | byte
}

type Integer int

func ReleaseResource[R json.Marshaler](r R) {
	r.MarshalJSON()
}
