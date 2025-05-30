package archive

import (
	"path/filepath"
)

func FilterFunc(filters []string) func(string) bool {
	if len(filters) == 0 {
		return func(path string) bool {
			return true // Allow everything
		}
	}
	return func(path string) bool {
		base := filepath.Base(path)
		for _, pat := range filters {
			match, _ := filepath.Match(pat, base)
			if match {
				return true
			}
			// Also match full relative path
			match, _ = filepath.Match(pat, path)
			if match {
				return true
			}
		}
		return false
	}
}
