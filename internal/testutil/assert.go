package testutil

import "testing"

func True(t *testing.T, cond bool, msg string) {
	t.Helper()
	if !cond {
		t.Errorf("expected true: %s", msg)
	}
}

func False(t *testing.T, cond bool, msg string) {
	t.Helper()
	if cond {
		t.Errorf("expected false: %s", msg)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func Equal[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

func Greater(t *testing.T, got, than float64, msg string) {
	t.Helper()
	if !(got > than) {
		t.Errorf("%s: got %v, want > %v", msg, got, than)
	}
}

func NotEmptyStrings(t *testing.T, s []string, msg string) {
	t.Helper()
	if len(s) == 0 {
		t.Errorf("expected non-empty slice: %s", msg)
	}
}

func EmptyStrings(t *testing.T, s []string, msg string) {
	t.Helper()
	if len(s) != 0 {
		t.Errorf("expected empty slice, got %v: %s", s, msg)
	}
}

func LenInt(t *testing.T, n, want int, msg string) {
	t.Helper()
	if n != want {
		t.Errorf("%s: got len %d, want %d", msg, n, want)
	}
}
