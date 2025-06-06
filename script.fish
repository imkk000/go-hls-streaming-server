#!/usr/bin/env fish

set -l path $argv[1]
if test -z $path
    echo "Usage: script.fish <path_to_video>"
    exit 1
end

set -l filename (path basename $path)
set -l output (string trim --right --chars=(path extension $filename) $filename)

cd segments
mkdir $output
cd $output

ffmpeg -i $path -map 0:v -map 0:a -codec: copy \
  -start_number 0 -hls_time 10 -hls_list_size 0 \
  -hls_segment_filename "segment_%03d.ts" -f hls playlist.m3u8
ffmpeg -i $path -map 0:s -c:s webvtt subtitle.vtt
ffmpeg -i segment_005.ts -frames:v 1 -update 1 thumbnail.png
ffmpeg -i $path -map 0:v:1 -update 1 cover.jpg
