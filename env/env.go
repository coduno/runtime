package env

import "os"

func IsDevAppServer() bool {
	if os.Getenv("CODUNO_DEV") == "true" {
		return true
	}
	return false
}
