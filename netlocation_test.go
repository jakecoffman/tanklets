package tanklets

import (
	"testing"
	"log"
)

func Benchmark_LocationSerialize(b *testing.B) {
	location := Location{
		ID: 1,
		X: 2,
		Y: 3,
		Vx: 4,
		Vy: 5,
		Angle: 6,
		AngularVelocity: 7,
		Turret: 8,
	}

	// saves time by allocating the correct size up front
	buffer := make([]byte, 0, 31)

	for n := 0; n < b.N; n++ {
		location.Serialize(buffer)
	}
}

func Benchmark_LocationDeserialize(b *testing.B) {
	location := Location{
		ID: 1,
		X: 2,
		Y: 3,
		Vx: 4,
		Vy: 5,
		Angle: 6,
		AngularVelocity: 7,
		Turret: 8,
	}

	bits, err := location.Serialize(nil)
	if err != nil {
		log.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		location.Serialize(bits)
	}
}
