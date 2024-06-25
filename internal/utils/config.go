package utils

import "github.com/tomruk/finddirs-go"

var FindDirsConfig = finddirs.AppConfig{
	Subdir:      "kopyaship",
	SubdirCache: "cache",
}

const APIFallbackAddr = "127.0.0.1:56792"
