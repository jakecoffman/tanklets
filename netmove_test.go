package tanklets

import (
	"testing"
	"math"
)

func TestMove_MarshalBinary(t *testing.T) {
	m := Move{
		1,
		1,
		math.MaxFloat64,
		math.MaxFloat64,
	}
	bytes, err := m.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	m2 := Move{}
	if err = m2.UnmarshalBinary(bytes); err != nil {
		t.Fatal(err)
	}

	if m != m2 {
		t.Fatal("Not equal", m, m2)
	}
}