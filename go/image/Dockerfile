FROM golang:1.12.0-alpine3.9

ENV APPNAME ''

ADD bingo/default /
ADD bingo/display_output /

WORKDIR /

ENTRYPOINT ["/display_output"]