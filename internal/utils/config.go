package utils

import "github.com/karagenc/finddirs-go"

var FindDirsConfig = finddirs.AppConfig{
	Subdir:      "kopyat",
	SubdirCache: "cache",
}

const APIFallbackAddr = "127.0.0.1:56792"
