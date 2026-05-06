package validatex

import "testing"

func TestValidateEmail(t *testing.T) {
	t.Parallel()

	if !ValidateEmail("user@example.com") {
		t.Fatal("expected valid email")
	}

	if ValidateEmail("bad@") {
		t.Fatal("expected invalid email")
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	if !ValidatePassword("pass123") {
		t.Fatal("expected valid password")
	}

	if ValidatePassword("short") {
		t.Fatal("expected invalid password")
	}
}

func TestValidateEmailQuery(t *testing.T) {
	t.Parallel()

	if !ValidateEmailQuery("user@example.com") {
		t.Fatal("expected valid query")
	}

	if ValidateEmailQuery("bad query") {
		t.Fatal("expected invalid query")
	}
}
