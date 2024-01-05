package main

type List interface {
	Add(idx int, val any) error
	Append(val any)
	Delete(idx int) (any, error)
}
