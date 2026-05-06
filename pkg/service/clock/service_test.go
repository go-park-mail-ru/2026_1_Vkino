package clock

import "testing"

func TestNow(t *testing.T) {
	t.Parallel()

	c := New()
	now := c.Now()
	if now.IsZero() {
		t.Fatal("expected non-zero time")
	}
}
