package utils

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	RunningOnWindows = runtime.GOOS == "windows"
	RunningOnMacOS   = runtime.GOOS == "darwin"
)

var RunningOnGitHubActions = os.Getenv("GITHUB_ACTIONS") == "true"

func StripDriveLetter(path string) string {
	if StartsWithDriveLetter(path) {
		return path[2:]
	}
	return path
}

func StartsWithDriveLetter(path string) bool {
	return len(path) >= 2 && (path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z') && path[1] == ':'
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

var (
	r       = rand.New(rand.NewSource(time.Now().UnixNano()))
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func ConvertToPascalCase(name string) string {
	return convertToMixedCase(name, true)
}

func ConvertToCamelCase(name string) string {
	return convertToMixedCase(name, false)
}

func convertToMixedCase(name string, firstToUpper bool) string {
	newName := strings.Builder{}
	newName.Grow(len(name))

	i := 0
	if firstToUpper {
		newName.WriteByte(strings.ToUpper(string(name[i]))[0])
		i++
	}

	for ; i < len(name); i++ {
		b := name[i]
		if b == '-' || b == '_' || b == ' ' && i != len(name)-1 {
			newName.WriteByte(strings.ToUpper(string(name[i+1]))[0])
		} else {
			newName.WriteByte(b)
		}
	}
	return newName.String()
}
