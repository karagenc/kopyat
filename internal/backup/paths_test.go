package backup

import "testing"

func TestPathCollision(t *testing.T) {
	paths := []string{
		"/var/tmp/xyz",
		"/var/lib/xyz",
		"/var/lib",
	}

	err := checkPathCollision(paths)
	if err == nil {
		t.Fatal("non-nil error expected")
	}
}
