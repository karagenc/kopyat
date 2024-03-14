package ifile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomruk/kopyaship/utils"
)

func TestSyncthingWalkInCurrentProject(t *testing.T) {
	os.Remove("test_ifile")
	i, err := New("test_ifile", ModeSyncthing, true, utils.NewCLILogger())
	if err != nil {
		require.NoError(t, err)
	}

	path, err := filepath.Abs("..")
	require.NoError(t, err)

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	content, err := os.ReadFile("test_ifile")
	require.NoError(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		require.NotContains(t, line, "README.md")
		require.NotContains(t, line, "Makefile")
		require.NotContains(t, line, "cmd/kopyaship")
		require.NotContains(t, line, "scripts")
		require.NotContains(t, line, "LICENSE")
		require.NotContains(t, line, ".gitignore")
	}
}

// Ensure the root path doesn't appear in ifile.
func TestSyncthingWalkNoRootMatchesInCurrentProject(t *testing.T) {
	os.Remove("test_ifile")
	i, err := New("test_ifile", ModeSyncthing, true, utils.NewCLILogger())
	if err != nil {
		require.NoError(t, err)
	}

	path, err := filepath.Abs("..")
	require.NoError(t, err)

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	content, err := os.ReadFile("test_ifile")
	require.NoError(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		if line == endIndicator {
			break // Don't check the empty line after the end indicator. It would be a false positive.
		} else if line == "" {
			t.Fatal("empty line. this means root path is ignored.")
		}
	}
}

func TestResticWalkInCurrentProject(t *testing.T) {
	os.Remove("test_ifile")
	i, err := New("test_ifile", ModeRestic, true, utils.NewCLILogger())
	if err != nil {
		require.NoError(t, err)
	}

	path, err := filepath.Abs("..")
	require.NoError(t, err)

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	content, err := os.ReadFile("test_ifile")
	require.NoError(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		require.NotContains(t, line, "tmp")
		require.NotContains(t, line, "kopyaship.yml")
		require.NotContains(t, line, "test_file")
	}
}
