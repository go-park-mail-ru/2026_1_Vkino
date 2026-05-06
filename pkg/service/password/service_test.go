package password

import "testing"

func TestHashAndCompare(t *testing.T) {
	t.Parallel()

	svc := New()

	hash, err := svc.Hash("secret")
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}

	if err := svc.Compare(hash, "secret"); err != nil {
		t.Fatalf("Compare error: %v", err)
	}

	if err := svc.Compare(hash, "wrong"); err == nil {
		t.Fatal("expected compare error for wrong password")
	}
}
