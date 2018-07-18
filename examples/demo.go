package main

import (
	"github.com/gookit/ini"
	"fmt"
)

func main() {
	cfg, err := ini.LoadFiles("testdata/test.ini")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", cfg.Data())

	cfg.LoadStrings(`
age = 100
[sec1]
newK = newVal
some = change val
`)
	fmt.Printf("%v\n", cfg.Data())

}