package types

import (
	"testing"
	"fmt"
)

func TestWindow(t *testing.T) {


	window := NewWindow()

	fmt.Printf("Window: \n%s\n", window.ToPrettyJson())
}