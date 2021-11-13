package main

import (
	"fmt"
	"testing"
)

var Test1 = `
@RequestParam(name="haha",defaultValue=123)
`

func TestParseAnnotation(t *testing.T) {

	m,err := ParseAnnotation(Test1)
	if err != nil {
		panic(err)
	}
	if len(m) ==0 {
		fmt.Println("len 0")
	}else{
		fmt.Println(to2PrettyString(m))
	}

}
