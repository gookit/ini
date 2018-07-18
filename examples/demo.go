package main

import (
	"github.com/gookit/ini"
	"fmt"
)

// go run ./examples/demo.go
func main() {
	cfg, err := ini.LoadFiles("testdata/test.ini")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", cfg.Data())

	// load more, will override prev data by key
	cfg.LoadStrings(`
age = 100
[sec1]
newK = newVal
some = change val
`)
	fmt.Printf("%v\n", cfg.Data())

}
