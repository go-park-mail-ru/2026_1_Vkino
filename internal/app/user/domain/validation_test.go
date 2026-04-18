package domain

import "testing"

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		email    string
		password string
		want     bool
	}{
		{
			name:     "valid credentials",
			email:    "user@example.com",
			password: "qwerty1",
			want:     true,
		},
		{
			name:     "invalid email",
			email:    "user@example",
			password: "qwerty1",
			want:     false,
		},
		{
			name:     "invalid password",
			email:    "user@example.com",
			password: "short",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Validate(tt.email, tt.password); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{name: "valid", password: "qwerty1", want: true},
		{name: "too short", password: "qwe1", want: false},
		{name: "contains spaces", password: "qwerty 1", want: false},
		{name: "no digits", password: "qwertyasd", want: false},
		{name: "no letters", password: "1234567", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ValidatePassword(tt.password); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestValidateEmailQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "valid full email", query: "user@example.com", want: true},
		{name: "valid partial query", query: "example", want: true},
		{name: "trimmed query", query: "  user@  ", want: true},
		{name: "empty", query: "", want: false},
		{name: "only spaces", query: "   ", want: false},
		{name: "contains internal spaces", query: "user example", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ValidateEmailQuery(tt.query); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
