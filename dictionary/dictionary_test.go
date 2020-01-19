package dictionary

import (
	"fmt"
	"testing"
)

func TestBuildDictionary(t *testing.T) {
	d := NewEmptyDictionary()
	d.BuildFromDir("../data")
	fmt.Println(d)
}
