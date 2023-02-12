package grid

import (
	"testing"
)

func TestOutOfBoundsLookup(t *testing.T) {
	wg := CreateWorldGrid(2048, 1125, 3)
	cell := wg.FindCell(-90, 0)

	if cell != nil {
		t.Logf("Found cell for OOB coordinates: %v", cell)
		t.Fail()
	}
}

func TestInBoundsLookup(t *testing.T) {
	wg := CreateWorldGrid(2048, 1125, 3)
	cell := wg.FindCell(1482.100708, 222.917847)

	if cell == nil {
		t.Logf("Didn't find cell for in-bounds coordinates")
		t.Fail()
	}

	t.Logf("%v", cell)
}

func TestSlightlyOutOfBoundsLookup(t *testing.T) {
	wg := CreateWorldGrid(2048, 1125, 3)
	cell := wg.FindCell(8400.1416, 9999.9999)

	if cell != nil {
		t.Logf("Found cell for slightly-OOB coordinates: %v", cell)
		t.Fail()
	}
}

func BenchmarkPointLookup(b *testing.B) {
	wg := CreateWorldGrid(2048, 1125, 3)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = wg.FindCell(8400.1416, 3211.1792)
	}
}
