# hlsXcoder (Go)

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

You can take any IP input and convert it to HLS. The container itself has it´s own CDN (https is by default) so you can serve the stream from the container. It has a built-in player that can show vumeters, subtitles (also DVB subs) and audios. It supports NVENC h264 or h265, deinterlace and resize. 

Create the Dockerfile.base container first, then build Dockerfile...

PD: certs are not valid for sure, use new ones.

Example:

docker run --runtime=nvidia --net=host --name channel_1 --restart unless-stopped --memory="256m" --memory-swap="256m" -d hlstranscoder:v1.4 \
./HLSTranscoder -i udp://239.192.21.6:2006 -ip 10.192.1.92 -port 8001 -hwaccel -cv h264_nvenc:500 -deint -ca aac -vumeter -r 480x270 -hls_list_size 10 -hls_time 3

Need help?:

docker run --runtime=nvidia --net=host --name channel_1 --restart unless-stopped --memory="256m" --memory-swap="256m" -d hlstranscoder:v1.4 \
./HLSTranscoder -h
