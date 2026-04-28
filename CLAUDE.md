# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```sh
go build -o server .        # build binary
./server --user <u> --pass <p>  # run (auth required by default)
./server --auth=false       # run without auth (dev only)
```

No task runner is present — use `go` directly. The binary is gitignored (`server`).

## Flags

| flag       | default    | notes                        |
| ---------- | ---------- | ---------------------------- |
| `user`     | —          | required when `--auth=true`  |
| `pass`     | —          | required when `--auth=true`  |
| `segments` | `segments` | path to HLS content root     |
| `port`     | `54321`    | binds to `127.0.0.1` only   |
| `auth`     | `true`     |                              |

## Architecture

Everything lives in `main.go` (single file). The server uses Echo with Basic Auth middleware. Templates are parsed at startup from `public/template/*.html`.

**Request flow:**

- `GET /playlist[?s=<path>]` — reads `segments/<path>` directory; for each subdirectory checks whether `playlist.m3u8` exists inside it. If yes → card links to `/?path=<subpath>` (player). If no → card links to `/playlist?s=<subpath>` (deeper browse). Passes `ParentURL` to template (empty string at root = no back button shown).
- `GET /?path=<path>` — renders the player page. Derives `ParentURL` via `filepath.Dir(path)`; empty if single-segment path.
- `GET /js/hls.render.js?path=<path>&v=<nano>` — dynamically generates JS that loads `/<path>/playlist.m3u8` into hls.js and attaches a subtitle track. The `v` param busts cache on each page load.
- `GET /progress?path=<path>` / `POST /progress?path=<path>` — in-memory playback resume. State is a `map[string]float64` guarded by `sync.RWMutex`. Resets on server restart. POST body: `{"t": <seconds>}`.

**Template data shapes:**

- `index.html` receives `Path` (string), `Timestamp` (int64), `ParentURL` (string)
- `list.html` receives `Dirs` ([]List), `ParentURL` (string)

**Security note:** the playlist handler validates that the resolved absolute path has the working directory as a prefix before serving, preventing path traversal.

## Content Preparation

`script.fish <segments_dir> <file.mkv> ...` converts MKV files to HLS using ffmpeg. Each input produces a subdirectory under `segments/` containing:
- `playlist.m3u8` (master, references `video.m3u8` + `audio.m3u8`)
- `video_NNN.ts` / `audio_NNN.ts` segments (10 s each)
- `subtitles.vtt`
- `cover.jpg` (thumbnail, 200 px height)

The server detects playable directories solely by the presence of `playlist.m3u8`.

## Frontend

No build step. Assets are served statically from `public/`:
- `public/js/` — `hls.min.js`, `plyr.js` (vendored)
- `public/css/` — `plyr.css`, `styles.css`
- `public/template/` — `index.html` (player), `list.html` (browser)

Styles are embedded directly in each template `<style>` block. The player page loads hls.js + Plyr, then the dynamically generated `hls.render.js`. Resume logic (fetch saved timestamp on load, POST every 5 s, `sendBeacon` on unload) is inlined in `index.html`.
