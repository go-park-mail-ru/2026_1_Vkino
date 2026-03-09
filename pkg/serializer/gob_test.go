package serializer

import (
	"errors"
	"testing"
)

type serializerTestUser struct {
	ID    int
	Email string
}

type serializerAnimal interface {
	Sound() string
}

type serializerDog struct {
	Name string
}

func (d serializerDog) Sound() string {
	return "woof"
}

type serializerZoo struct {
	Animal serializerAnimal
}

func TestSerialize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     any
		wantErrIs error
	}{
		{
			name: "success struct",
			value: serializerTestUser{
				ID:    1,
				Email: "user@example.com",
			},
		},
		{
			name:  "success string",
			value: "hello",
		},
		{
			name: "unregistered interface value",
			value: serializerZoo{
				Animal: serializerDog{Name: "Bob"},
			},
			wantErrIs: ErrSerialize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Serialize(tt.value)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("expected non-empty serialized data")
			}
		})
	}
}

func TestDeserialize(t *testing.T) {
	t.Parallel()

	validData, err := Serialize(serializerTestUser{
		ID:    7,
		Email: "test@example.com",
	})
	if err != nil {
		t.Fatalf("prepare valid data: %v", err)
	}

	tests := []struct {
		name      string
		data      []byte
		target    *serializerTestUser
		want      *serializerTestUser
		wantErrIs error
	}{
		{
			name:   "success",
			data:   validData,
			target: &serializerTestUser{},
			want: &serializerTestUser{
				ID:    7,
				Email: "test@example.com",
			},
		},
		{
			name:      "nil target",
			data:      validData,
			target:    nil,
			wantErrIs: ErrDeserialize,
		},
		{
			name:      "invalid bytes",
			data:      []byte("not gob data"),
			target:    &serializerTestUser{},
			wantErrIs: ErrDeserialize,
		},
		{
			name:      "empty data",
			data:      nil,
			target:    &serializerTestUser{},
			wantErrIs: ErrDeserialize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Deserialize(tt.data, tt.target)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.target == nil {
				t.Fatal("expected non-nil target")
			}
			if *tt.target != *tt.want {
				t.Fatalf("expected %+v, got %+v", *tt.want, *tt.target)
			}
		})
	}
}

func TestSerializeDeserialize_RoundTrip(t *testing.T) {
	t.Parallel()

	src := serializerTestUser{
		ID:    42,
		Email: "roundtrip@example.com",
	}

	data, err := Serialize(src)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	var dst serializerTestUser
	err = Deserialize(data, &dst)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}

	if dst != src {
		t.Fatalf("expected %+v, got %+v", src, dst)
	}
}

func TestRegisterType(t *testing.T) {
	zoo := serializerZoo{
		Animal: serializerDog{Name: "Rex"},
	}

	_, err := Serialize(zoo)
	if !errors.Is(err, ErrSerialize) {
		t.Fatalf("expected error %v before register, got %v", ErrSerialize, err)
	}

	RegisterType(serializerDog{})

	data, err := Serialize(zoo)
	if err != nil {
		t.Fatalf("serialize after register: %v", err)
	}

	var got serializerZoo
	err = Deserialize(data, &got)
	if err != nil {
		t.Fatalf("deserialize after register: %v", err)
	}

	if got.Animal == nil {
		t.Fatal("expected non-nil interface value")
	}

	dog, ok := got.Animal.(serializerDog)
	if !ok {
		t.Fatalf("expected serializerDog, got %T", got.Animal)
	}

	if dog.Name != "Rex" {
		t.Fatalf("expected dog name %q, got %q", "Rex", dog.Name)
	}
}
