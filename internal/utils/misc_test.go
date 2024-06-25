package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripDriveLetter(t *testing.T) {
	paths := []string{
		"A",
		":",
		":Z",
		"B:",
		"C:\\foo",
		"C:\\foo\\bar",
		"D:/foo",
		"D:/foo/bar",
	}
	expected := []string{
		"A",
		":",
		":Z",
		"",
		"\\foo",
		"\\foo\\bar",
		"/foo",
		"/foo/bar",
	}

	for i, path := range paths {
		newPath := StripDriveLetter(path)
		require.Equal(t, expected[i], newPath)
	}
}
