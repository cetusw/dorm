package individualtask

import (
	"github.com/google/uuid"
	"math"
	"strings"
	"testing"
	"time"
)

func TestValidation(t *testing.T) {
	if _, e := NormalizeTitle("  "); e == nil {
		t.Fatal("empty")
	}
	if _, e := NormalizeTitle(strings.Repeat("я", 255)); e != nil {
		t.Fatal(e)
	}
	if _, e := NormalizeTitle(strings.Repeat("я", 256)); e == nil {
		t.Fatal("long")
	}
	for _, v := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, e := NormalizeWeight(v); e == nil {
			t.Fatal(v)
		}
	}
	if v, e := NormalizeWeight(.04); e != nil || v != 0 {
		t.Fatal(v, e)
	}
}
func TestTransitions(t *testing.T) {
	n := time.Now()
	x, e := New(1, uuid.New(), nil, "x", 0, nil, n)
	if e != nil {
		t.Fatal(e)
	}
	if e = x.Complete(n); e != nil {
		t.Fatal(e)
	}
	if e = x.Reject(n); e != nil {
		t.Fatal(e)
	}
	_ = x.Complete(n)
	if e = x.Verify(n); e != nil {
		t.Fatal(e)
	}
	if e = x.Delete(n); e == nil {
		t.Fatal("verified delete")
	}
}
