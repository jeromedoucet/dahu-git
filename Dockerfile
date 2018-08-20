FROM alpine

LABEL maintainer Jerome Doucet <jerdct@gmail.com>

RUN apk --update add git openssh && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

RUN mkdir /data
RUN mkdir /app
WORKDIR /app

COPY dahu-git /app/

ENTRYPOINT ["./dahu-git"]
