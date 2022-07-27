FROM debian:latest
LABEL imgae.author="liuminhaw@gmail.com"

RUN apt update \
    && apt install -y wget \
    && wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && apt install -y ./google-chrome-stable_current_amd64.deb \
    && rm google-chrome-stable_current_amd64.deb \
    && apt autoclean -y \
    && rm -rf /var/lib/apt/lists/*

COPY loginCookie /usr/local/bin/