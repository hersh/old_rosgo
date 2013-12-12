package xmlrpc

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type Value struct {
	Int    *int     `xml:"int"`
	I4     *int     `xml:"i4"`
	String *string  `xml:"string"`
	Strval *string  `xml:",chardata"`
	Array  *[]Value `xml:"array>data>value"`
}

type Param struct {
	Val Value `xml:"value"`
}

type MethodCall struct {
	XMLName xml.Name `xml:"methodCall"`
	Name    string   `xml:"methodName"`
	Params  []Param  `xml:"params>param"`
}

type MethodResponse struct {
	XMLName xml.Name `xml:"methodResponse"`
	Params  []Param  `xml:"params>param"`
}

func (v Value) Print() {
	v.recursivePrint(0)
}

func (v Value) GetInt() (int, error) {
	switch {
	case v.Int != nil:
		return *v.Int, nil
	case v.I4 != nil:
		return *v.I4, nil
	default:
		return 0, errors.New("value does not contain an integer.")
	}
}

func (v Value) GetString() (string, error) {
	switch {
	case v.String != nil:
		return *v.String, nil
	case v.Strval != nil:
		return *v.Strval, nil
	default:
		return "", errors.New("value does not contain a string.")
	}
}

func (v Value) GetArray() []Value {
	switch {
	case v.Array != nil:
		return *v.Array
	default:
		return []Value{}
	}
}

func (v Value) recursivePrint(level int) {
	for i := 0; i < level; i++ {
		fmt.Printf("  ")
	}
	switch {
	case v.Int != nil:
		fmt.Printf("int: %d\n", *v.Int)
	case v.I4 != nil:
		fmt.Printf("i4: %d\n", *v.I4)
	case v.String != nil:
		fmt.Printf("string: %s\n", *v.String)
	case v.Strval != nil:
		fmt.Printf("strval: %s\n", *v.Strval)
	case v.Array != nil:
		array := *v.Array
		fmt.Printf("array of %d items:\n", len(array))
		for _, item := range array {
			item.recursivePrint(level + 1)
		}
	default:
		fmt.Printf("Empty Value\n")
	}
}
