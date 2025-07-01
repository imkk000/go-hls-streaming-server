#!/usr/bin/env fish

if test -z $argv[1]
    echo "Usage: script.fish <path_to_segments> <path_to_video> <path_to_video> ..."
    exit 1
end

set base_path $argv[1]

for path in $argv[2..-1]
    set filename (path basename $path)
    set output $base_path/(string trim --right --chars=(path extension $filename) $filename)

    mkdir -p $output

    ffmpeg -i $path -map 0:v -map 0:a -codec: copy \
        -start_number 0 -hls_time 10 -hls_list_size 0 \
        -hls_segment_filename $output/"segment_%03d.ts" -f hls $output/playlist.m3u8
    ffmpeg -i $path -map 0:s -c:s webvtt $output/subtitle.vtt
    ffmpeg -i $path -map 0:v:1 -update 1 $output/cover.jpg
end
