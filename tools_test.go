package toolkit

import "testing"

func TestTools_RandomString(t *testing.T) {
	var tools Tools

	s := tools.RandomString(10)
	if len(s) != 10 {
		t.Errorf("RandomString() = %v; want 10", len(s))
	}
}
