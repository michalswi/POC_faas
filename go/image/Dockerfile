# GO: display output

# GOOS=linux CGO_ENABLED=0 go build -a -ldflags '-w -s' -installsuffix cgo -o default default.go
# GOOS=linux CGO_ENABLED=0 go build -a -ldflags '-w -s' -installsuffix cgo -o display_output display_output.go
# GOOS=linux CGO_ENABLED=0 go build -a -ldflags '-w -s' -installsuffix cgo -o test test.go
# docker build -t local/go_faas:0.0.1 .
# docker run -v $(realpath test):/test -e APPNAME=test -p 8000:8000 local/go_faas:0.0.1
# docker run -v $(realpath test):/test -e APPNAME=test -e PORT=1234 -p 5050:1234 local/go_faas:0.0.1

FROM google/debian:wheezy

ARG GOBIN=go1.10.2.linux-amd64.tar.gz
COPY $GOBIN /tmp/$GOBIN
RUN tar -xzf /tmp/go1.10.2.linux-amd64.tar.gz -C /usr/local &&\
    rm -rf /tmp/$GOBIN

ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH $HOME/go
ENV APPNAME ''

ADD default ./default
ADD display_output ./display_output

ENTRYPOINT [ "/display_output" ]