package ifile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSyncthingWalkInCurrentProject(t *testing.T) {
	testIfile := testIfile("syncthing_walk_in_current_project")
	os.Remove(testIfile)
	i, err := New(testIfile, ModeSyncthing, true, zap.NewNop())
	if err != nil {
		require.NoError(t, err)
	}

	path, err := filepath.Abs("..")
	require.NoError(t, err)

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	content, err := os.ReadFile(testIfile)
	require.NoError(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		require.NotContains(t, line, "README.md")
		require.NotContains(t, line, "Makefile")
		require.NotContains(t, line, "cmd/kopyaship")
		require.NotContains(t, line, "scripts")
		require.NotContains(t, line, "LICENSE")
		require.NotRegexp(t, regexp.MustCompile("^/.gitignore$"), line)
	}
}

// Ensure the root path doesn't appear in ifile.
func TestSyncthingWalkNoRootMatchesInCurrentProject(t *testing.T) {
	testIfile := testIfile("syncthing_walk_no_root_matches_in_current_project")
	os.Remove(testIfile)
	i, err := New(testIfile, ModeSyncthing, true, zap.NewNop())
	if err != nil {
		require.NoError(t, err)
	}

	path, err := filepath.Abs("..")
	require.NoError(t, err)

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	content, err := os.ReadFile(testIfile)
	require.NoError(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		if line == endIndicator {
			break // Don't check the empty line after the end indicator. It would be a false positive.
		} else if line == "" {
			t.Fatal("empty line. this means root path is ignored.")
		}
	}
}

func TestSyncthingWalkInCurrentProjectAppend(t *testing.T) {
	testIfile := testIfile("syncthing_walk_in_current_project_append")
	os.Remove(testIfile)
	i, err := New(testIfile, ModeSyncthing, true, zap.NewNop())
	if err != nil {
		require.NoError(t, err)
	}

	path, err := filepath.Abs("..")
	require.NoError(t, err)

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	i, err = New(testIfile, ModeSyncthing, true, zap.NewNop())
	if err != nil {
		require.NoError(t, err)
	}

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	content, err := os.ReadFile(testIfile)
	require.NoError(t, err)

	found1 := true
	found2 := true
	for _, line := range strings.Split(string(content), "\n") {
		find1 := filepath.Join(path, "cmd/kopyaship/watch.go")
		if line == find1 {
			if found1 {
				t.Fatalf("this occurs more than once: %s", find1)
			}
			found1 = true
		}
		find2 := filepath.Join(path, "README.md")
		if line == find2 {
			if found2 {
				t.Fatalf("this occurs more than once: %s", find2)
			}
			found2 = true
		}
	}

	require.True(t, found1)
	require.True(t, found2)
}

func TestResticWalkInCurrentProject(t *testing.T) {
	testIfile := testIfile("restic_walk_in_current_project")
	os.Remove(testIfile)
	i, err := New(testIfile, ModeRestic, false, zap.NewNop())
	if err != nil {
		require.NoError(t, err)
	}

	path, err := filepath.Abs("..")
	require.NoError(t, err)

	err = i.Walk(path)
	require.NoError(t, err)
	err = i.Close()
	require.NoError(t, err)

	content, err := os.ReadFile(testIfile)
	require.NoError(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		require.NotContains(t, line, "tmp")
		require.NotContains(t, line, "kopyaship.yml")
		require.NotContains(t, line, "test_file")
	}
}
