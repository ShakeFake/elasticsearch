package utils

import (
	"fmt"
	"testing"
)

var (
	Path = "ab"
)

func TestReadFromPath(t *testing.T) {
	allCaption := ReadFromPath(Path)
	for _, caption := range allCaption {
		fmt.Println(caption)
	}
}
