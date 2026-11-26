// Package web holds a copy of the frontend embedded into the Go binary.
//
// The real source lives in ../frontend. The build step (see vercel.json)
// copies it into web/dist/ before `go build`, so the deployed binary can
// serve the page without any files on disk.
//
// web/dist/ is git-ignored except for .keep, which keeps `go build`
// working locally even when the folder has not been filled.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Files returns the embedded frontend, rooted at dist/.
func Files() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err) // only possible if the embed pattern above is wrong
	}
	return sub
}

// HasIndex reports whether index.html was embedded (i.e. the copy step ran).
func HasIndex() bool {
	_, err := fs.Stat(Files(), "index.html")
	return err == nil
}
