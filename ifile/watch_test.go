package ifile

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tomruk/kopyaship/utils"
)

func TestWatch(t *testing.T) {
	os.Remove("test_ifile")
	os.Remove("test_txtfile")

	defer os.Remove("test_txtfile")

	path, err := filepath.Abs("..")
	require.NoError(t, err)
	j := NewWatchJob(utils.NewCLILogger(true), path, "test_ifile", ModeSyncthing)

	var (
		walk      = j.walk
		content   []byte
		walkCount = 0
		mu        sync.Mutex
	)

	j.walk = func() error {
		walkErr := walk()
		c, err := os.ReadFile("test_ifile")

		mu.Lock()
		content = c
		walkCount++
		mu.Unlock()

		if !os.IsNotExist(err) {
			require.NoError(t, err)
		}
		return walkErr
	}

	go func() {
		err := j.Run()
		require.NoError(t, err)
	}()

	for {
		if j.Status() == WatchJobStatusRunning {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}

	err = os.WriteFile("test_txtfile", []byte(""), 0644)
	require.NoError(t, err)

	for {
		mu.Lock()
		if walkCount >= 1 {
			mu.Unlock()
			break
		}
		mu.Unlock()
		time.Sleep(time.Millisecond * 50)
	}

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
	os.Remove("test_txtfile")
	os.Remove(".gitignore")

	defer os.Remove("test_txtfile")
	defer os.Remove(".gitignore")

	path, err := filepath.Abs("..")
	require.NoError(t, err)
	j := NewWatchJob(utils.NewCLILogger(true), path, "test_ifile", ModeSyncthing)

	var (
		walk      = j.walk
		content   []byte
		walkCount = 0
		mu        sync.Mutex
	)

	j.walk = func() error {
		walkErr := walk()
		c, err := os.ReadFile("test_ifile")

		mu.Lock()
		content = c
		walkCount++
		mu.Unlock()

		if !os.IsNotExist(err) {
			require.NoError(t, err)
		}
		return walkErr
	}
	go func() {
		err := j.Run()
		require.NoError(t, err)
	}()

	for {
		if j.Status() == WatchJobStatusRunning {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	err = os.WriteFile(".gitignore", []byte("test_txtfile"), 0644)
	require.NoError(t, err)

	err = os.WriteFile("test_txtfile", []byte(""), 0644)
	require.NoError(t, err)

	for {
		mu.Lock()
		if walkCount >= 2 {
			mu.Unlock()
			break
		}
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}

	found := false
	mu.Lock()
	for _, line := range strings.Split(string(content), "\n") {
		entry := "/ifile/test_txtfile"
		if line == entry {
			found = true
		}
	}
	mu.Unlock()
	require.True(t, found)

	err = j.Shutdown()
	require.NoError(t, err)
}
