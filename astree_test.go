package main

import (
	"fmt"
	"testing"
)

func TestAst1(t *testing.T) {
	var scheduleIds map[int64]string = make(map[int64]string)
	scheduleIds[1] = "a"
	scheduleIds[2] = "b"

	delete(scheduleIds, 3)
	delete(scheduleIds, 1)
	for k, v := range scheduleIds {
		fmt.Println(k, v)
	}
}
