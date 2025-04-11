# see compose.yml

FROM debian:bookworm

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        build-essential golang && \
    rm -rf /var/lib/apt/lists

RUN useradd --create-home --shell /bin/bash ctf

COPY kv.go frontend.go flag.txt /home/ctf/
WORKDIR /root
RUN cd /home/ctf && \
    go build -asan kv.go && \
    go build -asan frontend.go && \
    rm *.go

RUN chmod 555 /home/ctf && \
    chown -R root:root /home/ctf && \
    chmod -R 000 /home/ctf/* && \
    chmod 444 /home/ctf/flag.txt && \
    chmod 005 /home/ctf/kv /home/ctf/frontend

RUN find / -ignore_readdir_race -path /sys -prune -o -type f \( -perm -4000 -o -perm -2000 \) -not -wholename /readflag -exec rm {} \;
USER ctf
RUN ! find / -writable -or -user $(id -un) -or -group $(id -Gn|sed -e 's/ / -or -group /g') 2> /dev/null | grep -Ev -m 1 '^(/dev/|/run/|/proc/|/sys/|/tmp|/var/tmp|/var/lock)'
USER root


WORKDIR /home/ctf
EXPOSE 1024
USER ctf
CMD while true; do sleep 1m; find /tmp -mindepth 1 -mmin '+10' -delete; done & \
    /home/ctf/frontend
