FROM ubuntu:22.04

WORKDIR /usr/src/app

COPY wheatleycrab .

RUN apt-get update && apt-get install -y ffmpeg youtube-dl

CMD ["./wheatleycrab"]
