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

    ffmpeg -i $path -map 0:s -c:s webvtt $output/subtitles.vtt &
    ffmpeg -i $path -map 0:v:1 -update 1 -vf scale=-1:200 $output/cover.jpg &
    ffmpeg -threads 0 -i $path -map 0:v:0 -c:v libx264 \
        -profile:v main -level 3.1 -pix_fmt yuv420p \
        -an -sn -f hls -start_number 0 \
        -hls_segment_filename $output/"video_%03d.ts" \
        -hls_time 10 -hls_playlist_type vod $output/video.m3u8 &
    ffmpeg -threads 0 -i $path -map 0:a:0 -c:a aac -vn -ac 2 -f hls \
        -start_number 0 -hls_flags independent_segments \
        -hls_segment_filename $output/"audio_%03d.ts" \
        -hls_time 10 -hls_playlist_type vod $output/audio.m3u8 &

    wait

    echo '#EXTM3U
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="Main",DEFAULT=YES,URI="audio.m3u8"
#EXT-X-STREAM-INF:BANDWIDTH=5000000,AUDIO="audio"
video.m3u8' >$output/playlist.m3u8
end
