package ifile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tomruk/kopyaship/utils"
)

func TestWatch(t *testing.T) {
	os.Remove("test_ifile")

	path, err := filepath.Abs("..")
	require.NoError(t, err)
	j := NewWatchJob(utils.NewCLILogger(), path, "test_ifile", ModeSyncthing)
	go func() {
		err := j.Run()
		require.NoError(t, err)
	}()

	time.Sleep(2 * time.Second)

	err = os.WriteFile("test_txtfile", []byte(""), 0644)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)
	content, err := os.ReadFile("test_ifile")
	require.NoError(t, err)

	for _, line := range strings.Split(string(content), "\n") {
		entry := "/ifile/test_txtfile"
		if line == entry {
			t.Fatalf("this shouldn't have been in ifile: %s", entry)
		}
	}

	err = j.Shutdown()
	require.NoError(t, err)
}

// Make sure newly created .gitignore and ignored file gets watched and entry gets added.
func TestWatchIgnore(t *testing.T) {
	os.Remove("test_ifile")

	path, err := filepath.Abs("..")
	require.NoError(t, err)
	j := NewWatchJob(utils.NewCLILogger(), path, "test_ifile", ModeSyncthing)
	go func() {
		err := j.Run()
		require.NoError(t, err)
	}()

	time.Sleep(2 * time.Second)

	err = os.WriteFile(".gitignore", []byte("test_txtfile"), 0644)
	defer os.Remove(".gitignore")
	require.NoError(t, err)

	err = os.WriteFile("test_txtfile", []byte(""), 0644)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)
	content, err := os.ReadFile("test_ifile")
	require.NoError(t, err)

	found := false
	for _, line := range strings.Split(string(content), "\n") {
		entry := "/ifile/test_txtfile"
		if line == entry {
			found = true
		}
	}
	require.True(t, found)

	err = j.Shutdown()
	require.NoError(t, err)
}
