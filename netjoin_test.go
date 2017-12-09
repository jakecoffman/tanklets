package tanklets

import (
	"testing"
	"math"
	"github.com/go-gl/mathgl/mgl32"
)

func TestJoin_MarshalBinary(t *testing.T) {
	j := Join {
		PlayerID(math.MaxUint16),
		true,
		mgl32.Vec3{1, 2, 3},
	}
	bytes, err := j.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	j2 := Join{}
	err = j2.UnmarshalBinary(bytes)
	if err != nil {
		t.Fatal(err)
	}

	if j != j2 {
		t.Error("joins not equal", j, j2)
	}
}
