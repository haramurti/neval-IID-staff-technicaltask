package app

import (
	"log"
	"net/http"
	"sync"
)

// Serverless platforms (Vercel) call a handler function per request and
// may reuse the same process for many requests. We build the App only once
// per process (sync.Once), then reuse it.
var (
	sharedOnce sync.Once
	sharedApp  *App
	sharedErr  error
)

// Shared returns the process-wide App, created from env on first use.
func Shared() (*App, error) {
	sharedOnce.Do(func() {
		sharedApp, sharedErr = NewFromEnv()
		if sharedErr != nil {
			log.Printf("startup: %v", sharedErr)
		}
	})
	return sharedApp, sharedErr
}

// StartupFailed writes a generic 500 when config is broken (details are logged).
func StartupFailed(w http.ResponseWriter) {
	http.Error(w, "server is not configured correctly", http.StatusInternalServerError)
}
