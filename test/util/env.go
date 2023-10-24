package util

import "os"

func envOr(name, def string) string {
	if val, ok := os.LookupEnv(name); ok {
		return val
	} else {
		return def
	}
}
