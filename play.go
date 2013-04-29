package main

import (
	"fmt"
)

type A struct {
	Ayy string
}

func (a A) Show() {
	fmt.Println( a.Ayy )
}

type B struct {
	A
	Bee int
}

func (b B) Count() {
	fmt.Println( b.Bee )
}

func main() {
	b := new( B )
	b.Ayy = "foo"
	b.Bee = 17
	fmt.Println( b )
	b.Count()
	b.Show()
}
