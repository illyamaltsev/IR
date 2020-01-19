package main

import (
	"InfoPoisk/dictionary"
)

func main() {
	d := dictionary.NewEmptyDictionary()
	d.FillFromTxt("kobzar.txt")
	print(d)
	//d.save("dict.txt")
}
