package gofast

import "compress/flate"
import s "github.com/prataprc/gosettings"

func DefaultSettings(start, end int64) s.Settings {
	return s.Settings{
		"buffersize":   512,
		"batchsize":    1,
		"chansize":     100000,
		"tags":         "",
		"opaque.start": start,
		"opaque.end":   end,
		"log.level":    "info",
		"log.file":     "",
		"gzip.level":   flate.BestSpeed,
	}
}
