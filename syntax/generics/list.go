package main

import "errors"

// T 类型参数，名字叫做 T，约束是 any，等于没有约束
type List[T any] interface {
	Add(idx int, t T)
	Append(t T)
}

//type ListFloat interface {
//	Add(idx int, t float)
//	Append(t float)
//}

func main() {
	//UseList()
	println(Sum[int](1, 2, 3))
	println(Sum[Integer](1, 2, 3))
	println(Sum[float64](1.1, 2.1, 3.1))
	println(Sum[float32](1.1, 2.1, 3.1))
	var j MyMarshal
	ReleaseResource[*MyMarshal](&j)
}

type MyMarshal struct {
}

func (m *MyMarshal) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func UseList() {
	//var l List[int]
	//l.Append(12)

	var lany List[any]
	lany.Append(12.3)
	lany.Append(123)
	lk := LinkedList[int]{}
	intVal := lk.head.val
	println(intVal)
}

// type parameter
type LinkedList[Daming any] struct {
	head *node[Daming]
	t    Daming
	tp   *Daming
	tp2  ***Daming
}

type node[T any] struct {
	val T
}

func Max[T Number](vals ...T) (T, error) {
	if len(vals) == 0 {
		var t T
		return t, errors.New("你的下标不对")
	}
	res := vals[0]
	for i := 1; i < len(vals); i++ {
		if res < vals[i] {
			res = vals[i]
		}
	}
	return res, nil
}

func AddSlice[T any](slice []T, idx int, val T) ([]T, error) {
	// 如果我这边 idx 是负数，或者超过了 slice 的界限
	if idx < 0 || idx > len(slice) {
		return nil, errors.New("下标出错")
	}

	res := make([]T, 0, len(slice)+1)
	for i := 0; i < idx; i++ {
		res = append(res, slice[i])
	}
	res[idx] = val
	for i := idx; i < len(slice); i++ {
		res = append(res, slice[i])
	}

	return res, nil
}
