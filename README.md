# Home Streaming Server

> **Disclaimer:** Parts of the code in this project were generated with assistance from [Claude Code](https://claude.ai/code) (Anthropic). The core architecture, design decisions, and overall implementation are my own work. Claude Code was used to assist with specific features, not to generate the full codebase.

## Why?

I want a simple personal home streaming server. No external provider needed — I just want to watch movies at home.
Secured with `basic auth` and exposed safely via `cloudflared` without revealing my public IP.
Supports picture-in-picture mode, so it works well on iPad, mobile phone, and browser.

## Features

- HLS video streaming with subtitle support
- Basic auth protected
- Nested directory browsing (series → season → episode)
- Cover image cards for each title
- Parent directory navigation (back button)
- Resume playback — timestamp saved in-memory on the server, so you can continue where you left off or start from the beginning

## Directory Structure

Source files are pre-converted MKV files containing video, English subtitle, and cover (JPG).  
Each MKV is converted to a `playlist.m3u8` (with `video.m3u8`, `audio.m3u8`), `subtitles.vtt`, and segmented `.ts` files.

Run `script.fish movie1.mkv movie2.mkv ...` to convert and place into the `segments` folder.

```
segments
├── movie1
│   ├── playlist.m3u8
│   ├── video.m3u8
│   ├── audio.m3u8
│   ├── subtitles.vtt
│   └── cover.jpg
└── series
    ├── season1
    │   ├── ep1
    │   │   ├── playlist.m3u8
    │   │   └── cover.jpg
    │   └── ep2
    │       ├── playlist.m3u8
    │       └── cover.jpg
    └── season2
        └── ep1
            ├── playlist.m3u8
            └── cover.jpg
```

## Stack

- **Backend:** Go, [Echo](https://echo.labstack.com/), [zerolog](https://github.com/rs/zerolog)
- **Frontend:** Vanilla HTML/CSS/JS, [Plyr](https://plyr.io/), [hls.js](https://github.com/video-dev/hls.js/)
- **Tunnel:** [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/)

## Server Flags

```sh
# build
go build -o server .

# run locally (uses 127.0.0.1)
./server --user <user> --pass <pass>

# run without auth (dev only)
./server --auth=false
```

| flag       | description             | required | default    |
| ---------- | ----------------------- | -------- | ---------- |
| `user`     | basic auth username     | ✅       |            |
| `pass`     | basic auth password     | ✅       |            |
| `segments` | path to segments folder |          | `segments` |
| `port`     | server port             |          | `54321`    |
| `auth`     | enable basic auth       |          | `true`     |

## API

| method | path                                  | description                                    |
| ------ | ------------------------------------- | ---------------------------------------------- |
| GET    | `/playlist`                           | root directory browser                         |
| GET    | `/playlist?s=<path>`                  | subdirectory browser                           |
| GET    | `/?path=<path>`                       | video player for a given path                  |
| GET    | `/js/hls.render.js?path=<path>&v=<n>` | dynamically generated HLS + subtitle loader JS |
| GET    | `/progress?path=<path>`               | get saved playback timestamp                   |
| POST   | `/progress?path=<path>`               | save playback timestamp `{"t": 123}`           |

> Note: playback timestamps are stored in-memory only and reset when the server restarts. On video end, the timestamp is reset to 0.
