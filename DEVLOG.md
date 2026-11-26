# DEVLOG

Real problems hit while building, with cause and fix. Newest first.

---

## 5. No EXPA access token before the deadline

**When:** whole project (planning on 26 Nov 2026).

**Problem:** the task needs LC numbers via GraphQL from EXPA, but no access
token was available. Without a token the EXPA schema cannot even be inspected,
so the real query could not be written.

**Fix:** a repository interface between the service and the data source.
`MockRepository` serves realistic, reproducible seed data. `ExpaRepository` is a
marked stub with a step-by-step TODO, and does not guess EXPA field names. The
switch is one env var (`DATA_SOURCE`). The UI badge and the CSV say "mock" so
dummy numbers are never presented as real.

---

## 4. Vercel build failed: "No compiled Go binary found after buildCommand"

**When:** first deploy to Vercel (26 Nov 2026).

**Error:**
```
Using Go 1.26.0 (from backend/go.mod)
Error: [BUILD_FAILED] No compiled Go binary found after buildCommand.
Ensure your command produces a Linux binary (GOOS=linux is set),
or use `go build -o $VERCEL_OUTPUT_FILE`.
```

**Cause:** The plan was Vercel's classic Go setup: one serverless function per
file in `api/`, plus the frontend copied to `public/`. But because the project
root (`backend/`) has a `go.mod`, Vercel detected it as a **Go backend
framework** project. In that mode Vercel ignores `api/` and expects the build
command to produce **one server binary** that handles every request. Our build
command only copied files, so no binary was found.

**Fix:** Switched to the framework mode, which turned out simpler:
- `vercel.json` build command now runs `go build -o $VERCEL_OUTPUT_FILE ./cmd/server`,
  the same server we run locally.
- The deployed binary has no `../frontend` folder next to it, so the build
  first copies `frontend/` into `backend/web/dist/` and the binary **embeds** it
  with `//go:embed`. Locally the server still reads `../frontend` from disk, so
  frontend edits show on refresh.
- Removed `api/*.go`, the `/graphql` rewrites and the `sync.Once` helper. They
  were no longer needed.

**Verified:** ran the exact build command with `GOOS=linux`, then ran the binary
from a folder with no `frontend/`. `/`, `/app.js`, `/styles.css`, `/graphql`
and `/export.csv` all returned 200 from the embedded copy.

**Lesson:** read how the platform *detects* the project type, not just how it
runs code. One `go.mod` changed Vercel's whole build mode.

---

## 3. gqlgen generated a GraphQL error for a request the browser handles fine

**When:** smoke-testing the HTTP handlers with `httptest`.

**Cause:** gqlgen's POST transport only accepts `Content-Type: application/json`.
The test request had no header, so gqlgen rejected it. The browser `fetch` in
`app.js` already sets the header, so users were never affected.

**Fix:** added the header in the test. Worth knowing for anyone calling the API
with `curl`: always pass `-H 'Content-Type: application/json'`.

---

## 2. Latest gqlgen requires Go 1.26

**When:** installing gqlgen (`go get -tool github.com/99designs/gqlgen@latest`).

**Cause:** the latest gqlgen release needs Go 1.26, so `go get` bumped the
`go` line in `go.mod` from 1.25 to 1.26. The local Go (1.25.5) downloads the
1.26 toolchain automatically, so this was silent.

**Risk / fix:** a deploy platform might not support 1.26. Vercel downloaded Go
1.26.0 from `go.mod` without problems, so no change was needed. The fallback
was to pin an older gqlgen version.

---

## 1. Styles "missing" after adding the status box

**When:** testing the loading / empty / error states.

**Cause:** the browser kept an old cached `styles.css`, so the new `.status`
classes did not exist on the page. The file served by the server was correct.

**Fix:** hard refresh (Cmd+Shift+R). Lesson: when CSS changes "don't work",
check the served file first before changing code.
