package validator

import (
	"reflect"
	"testing"
)

func TestIsQueryValueCompatible(t *testing.T) {
	if !isQueryValueCompatible("abc", reflect.TypeFor[string]()) {
		t.Fatalf("string should be compatible")
	}
	if !isQueryValueCompatible("true", reflect.TypeFor[bool]()) {
		t.Fatalf("bool should be compatible")
	}
	if isQueryValueCompatible("x", reflect.TypeFor[bool]()) {
		t.Fatalf("invalid bool should be incompatible")
	}
	if !isQueryValueCompatible("10", reflect.TypeFor[int64]()) {
		t.Fatalf("int should be compatible")
	}
	if isQueryValueCompatible("x", reflect.TypeFor[int64]()) {
		t.Fatalf("invalid int should be incompatible")
	}
	if !isQueryValueCompatible("10", reflect.TypeFor[uint64]()) {
		t.Fatalf("uint should be compatible")
	}
	if isQueryValueCompatible("-1", reflect.TypeFor[uint64]()) {
		t.Fatalf("negative uint should be incompatible")
	}
	if !isQueryValueCompatible("1.23", reflect.TypeFor[float64]()) {
		t.Fatalf("float should be compatible")
	}
	if isQueryValueCompatible("x", reflect.TypeFor[float64]()) {
		t.Fatalf("invalid float should be incompatible")
	}
	if !isQueryValueCompatible("anything", reflect.TypeFor[struct{}]()) {
		t.Fatalf("struct should fallback to compatible")
	}
}

func TestExpectedTypeLabel(t *testing.T) {
	cases := []struct {
		in   reflect.Type
		want string
	}{
		{reflect.TypeFor[string](), "string"},
		{reflect.TypeFor[bool](), "boolean"},
		{reflect.TypeFor[int](), "number"},
		{reflect.TypeFor[uint](), "number"},
		{reflect.TypeFor[float32](), "number"},
		{reflect.TypeFor[struct{}](), "struct"},
	}

	for _, tc := range cases {
		if got := expectedTypeLabel(tc.in); got != tc.want {
			t.Fatalf("expected %q, got %q", tc.want, got)
		}
	}
}

func TestNormalizeJSONField(t *testing.T) {
	if got := normalizeJSONField(""); got != "request_body" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeJSONField("payload.user.name"); got != "name" {
		t.Fatalf("got %q", got)
	}
}
