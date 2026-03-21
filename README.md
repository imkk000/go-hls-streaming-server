# Home Streaming Server

## Why?

I want to implement simple own home web streaming server. No need other provider, because I just need to watch a movie.
Require `basic auth` username and password. I use `cloudflared` to online without expose public IP.

## How?

I store my source movie into mkv file. There are video, subtitle (English), and cover (JPG).
I convert the mkv file into `playlist.m3u8` which include `video.m3u8` and `audio.m3u8`, and also convert subtitle to `subtitles.vtt`.
The video segments in format `video_n.ts`, and audio segments in format `audio_n.ts`.
I run `script.fish movie1.mkv movie2.mkv ...` to pre-convert what movie I want to watch then write into default folders `segments`.
For series or movie collections, my server support nested directory too.

```
segments
├── movie1 -> playlist
└── series
    ├── season1
    │   ├── ep1 -> playlist
    │   └── ep2 -> playlist
    └── season2
        └── ep1 -> playlist
```

## Server Flags

```sh
# build
task build

# deploy to systemd with my user (rootless)
# create media.service into ~/.config/systemd/user/
task deploy

# use 127.0.0.1 instead of 0.0.0.0
./server --user <user> --pass <pass>
```

| name     | description             | required | default  |
| -------- | ----------------------- | -------- | -------- |
| user     | set basic auth username | ✅       |          |
| pass     | set basic auth password | ✅       |          |
| segments | set segments path       |          | segments |
| port     | set server port         |          | 54321    |
