package main

import (
	"fmt"

	"./dictionary"
)

func main() {
	d := dictionary.NewEmptyDictionary()
	d.BuildFromDir("data")
	fmt.Println(d)
	//d.save("dict.txt")
}
