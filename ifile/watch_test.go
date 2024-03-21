package ifile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tomruk/kopyaship/utils"
	"go.uber.org/zap"
)

var scanPath = func() string {
	f, err := filepath.Abs("..")
	if err != nil {
		panic(err)
	}
	return f
}()

func TestWatch(t *testing.T) {
	const testTxtfile = "test_txtfile_watch"
	testIfile := testIfile("watch")
	os.Remove(testIfile)
	os.Remove(testTxtfile)

	j := NewWatchJob(testIfile, scanPath, ModeSyncthing, nil, nil, zap.NewNop())

	var (
		walk      = j.walk
		content   []byte
		walkCount = 0
		mu        sync.Mutex
	)

	j.walk = func() error {
		walkErr := walk()
		c, err := os.ReadFile(testIfile)

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

	err := os.WriteFile(testIfile, []byte(""), 0644)
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
		entry := "/ifile/" + testTxtfile
		if line == entry {
			t.Fatalf("this shouldn't have been in ifile: %s", entry)
		}
	}

	err = j.Shutdown()
	require.NoError(t, err)
}

// Make sure newly created .gitignore and ignored file gets watched and entry gets added.
func TestWatchIgnore(t *testing.T) {
	const testTxtfile = "test_txtfile_watch_ignore"
	testIfile := testIfile("watch_ignore")
	os.Remove(testIfile)
	os.Remove(testTxtfile)
	os.Remove(".gitignore")

	defer os.Remove(testTxtfile)
	defer os.Remove(".gitignore")

	j := NewWatchJob(testIfile, scanPath, ModeSyncthing, nil, nil, zap.NewNop())

	var (
		walk      = j.walk
		content   []byte
		walkCount = 0
		mu        sync.Mutex
	)

	j.walk = func() error {
		walkErr := walk()
		c, err := os.ReadFile(testIfile)

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

	err := os.WriteFile(".gitignore", []byte(testTxtfile), 0644)
	require.NoError(t, err)

	err = os.WriteFile(testTxtfile, []byte(""), 0644)
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
		entry := "/ifile/" + testTxtfile
		if line == entry {
			found = true
		}
	}
	mu.Unlock()
	require.True(t, found)

	err = j.Shutdown()
	require.NoError(t, err)
}

// Make sure newly created .gitignore and ignored file inside newly created directory gets watched and entry gets added.
func TestWatchIgnoreNewlyCreatedDir(t *testing.T) {
	const (
		testDir     = "testdir_watch_ignore_newly_created_dir"
		testTxtfile = "test_txtfile_watch_ignore_newly_created_dir"
	)
	testIfile := testIfile("watch_ignore_newly_created_dir")
	os.Remove(testIfile)
	os.RemoveAll(testDir)
	os.Remove(".gitignore")

	defer os.RemoveAll(testDir)
	defer os.Remove(".gitignore")

	j := NewWatchJob(testIfile, scanPath, ModeSyncthing, nil, nil, zap.NewNop())

	var (
		walk      = j.walk
		content   []byte
		walkCount = 0
		mu        sync.Mutex
	)

	j.walk = func() error {
		walkErr := walk()
		c, err := os.ReadFile(testIfile)

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

	err := os.Mkdir(testDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(".gitignore", []byte("/"+testDir+"/"+testTxtfile), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(testDir, testTxtfile), []byte(""), 0644)
	require.NoError(t, err)

	for {
		mu.Lock()
		if walkCount >= 3 {
			mu.Unlock()
			break
		}
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
	}

	found := false
	mu.Lock()
	for _, line := range strings.Split(string(content), "\n") {
		entry := "/ifile/" + testDir + "/" + testTxtfile
		if line == entry {
			found = true
		}
	}
	mu.Unlock()
	require.True(t, found)

	err = j.Shutdown()
	require.NoError(t, err)
}

func TestWatchFail(t *testing.T) {
	const (
		testTxtfile1 = "test_txtfile_watchfail_1"
		testTxtfile2 = "test_txtfile_watchfail_2"
	)
	testIfile := testIfile("watch_fail")
	os.Remove(testIfile)
	os.Remove(testTxtfile1)
	os.Remove(testTxtfile2)

	j := NewWatchJob(testIfile, scanPath, ModeSyncthing, nil, nil, zap.NewNop())
	j.failAfter = 4

	var (
		walkCount = 0
		mu        sync.Mutex
	)

	j.walk = func() error {
		mu.Lock()
		walkCount++
		walkCount := walkCount - 1
		mu.Unlock()

		if walkCount == 0 {
			return nil
		}
		return fmt.Errorf("test walk error")
	}

	go func() {
		err := j.Run()
		require.Error(t, err)
	}()

	for {
		mu.Lock()
		if walkCount >= 1 {
			mu.Unlock()
			break
		}
		mu.Unlock()
		fmt.Println("Waiting for walkCount >= 1")
		time.Sleep(time.Millisecond * 50)
	}

	for {
		if j.Status() == WatchJobStatusRunning {
			break
		}
		fmt.Println("Waiting for j.Status() == WatchJobStatusRunning")
		time.Sleep(time.Millisecond * 50)
	}

	// Trigger walk
	if utils.RunningOnGitHubActions && utils.RunningOnMacOS {
		// Trigger walk manually.
		// For some reason, fsnotify events are not emitted on macOS GitHub Action runner.
		for {
			if v := j.testEventChanSender.Load(); v != nil {
				f := v.(func(string))
				absTestTxtfile1, err := filepath.Abs(testTxtfile1)
				require.NoError(t, err)
				f(absTestTxtfile1)
				break
			}
			fmt.Println("Waiting for j.testEventChanSender.Load() != nil")
			time.Sleep(50 * time.Millisecond)
		}
	} else {
		err := os.WriteFile(testTxtfile1, nil, 0644)
		require.NoError(t, err)
		err = os.Remove(testTxtfile1)
		require.NoError(t, err)
	}

	for {
		mu.Lock()
		if walkCount >= 2 {
			mu.Unlock()
			break
		}
		mu.Unlock()
		fmt.Println("Waiting for walkCount >= 2")
		time.Sleep(time.Millisecond * 50)
	}

	retries := 0
	for {
		info := j.Info()
		if len(info.Errors) > 0 {
			break
		}
		if retries >= 100 {
			t.Fatal("expected (waited for) len(info.Errors) to be greater than 0, but waiting timed out")
		}
		time.Sleep(time.Millisecond * 50)
		fmt.Println("Waiting for len(info.Errors) to be greater than 0")
		retries++
		continue
	}

	// Ensure 4 seconds (value of failAfter) has passed.
	time.Sleep(4010 * time.Millisecond)

	// Trigger walk again
	if utils.RunningOnGitHubActions && utils.RunningOnMacOS {
		// Trigger walk manually.
		// For some reason, fsnotify events are not emitted on macOS GitHub Action runner.
		for {
			if v := j.testEventChanSender.Load(); v != nil {
				f := v.(func(string))
				absTestTxtfile1, err := filepath.Abs(testTxtfile2)
				require.NoError(t, err)
				f(absTestTxtfile1)
				break
			}
			fmt.Println("Waiting for j.testEventChanSender.Load() != nil")
			time.Sleep(50 * time.Millisecond)
		}
	} else {
		err := os.WriteFile(testTxtfile2, nil, 0644)
		require.NoError(t, err)
		err = os.Remove(testTxtfile2)
		require.NoError(t, err)
	}

	for {
		if j.Status() == WatchJobStatusFailed {
			break
		}
		fmt.Println("Waiting for j.Status() == WatchJobStatusFailed")
		time.Sleep(time.Millisecond * 50)
	}

	retries = 0
	for {
		info := j.Info()
		if len(info.Errors) > 1 {
			break
		}
		if retries >= 100 {
			t.Fatal("expected (waited for) len(info.Errors) to be greater than 1, but waiting timed out")
		}
		time.Sleep(time.Millisecond * 50)
		fmt.Println("Waiting for len(info.Errors) to be greater than 1")
		retries++
		continue
	}

	err := j.Shutdown()
	require.NoError(t, err)
}

func TestWatchFailImmediately(t *testing.T) {
	testIfile := testIfile("watch_fail_immediately")
	os.Remove(testIfile)

	runHooks := func() error { return fmt.Errorf("nothing") } // Just so that coverage is triggered.
	j := NewWatchJob(testIfile, scanPath, ModeSyncthing, runHooks, runHooks, zap.NewNop())

	var (
		walkCount = 0
		mu        sync.Mutex
	)

	j.walk = func() error {
		mu.Lock()
		walkCount++
		mu.Unlock()
		return fmt.Errorf("test walk error")
	}

	err := j.Run()
	require.Error(t, err)

	require.Equal(t, j.Status(), WatchJobStatusFailed)

	require.Equal(t, j.Ifile(), j.ifile)       // Just so that coverage is triggered.
	require.Equal(t, j.ScanPath(), j.scanPath) // Just so that coverage is triggered.
}
