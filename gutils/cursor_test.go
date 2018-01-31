package gutils

import (
	"testing"
)

func TestCursor_Next(t *testing.T) {
	c := NewCursor(1, 2)
	v := c.Next()
	if v != 1 {
		t.Error("Should be 1 got ", v)
	}
	v = c.Next()
	if v != 2 {
		t.Error("Should be 2 got ", v)
	}
	v = c.Next()
	if v != 1 {
		t.Error("Should be 1 got ", v)
	}
}
