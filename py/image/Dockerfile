# PY: display output

# docker build -t local/py_faas:0.0.1 --build-arg http_p=$http_proxy .
# docker run -it -v $(realpath .)/z.py:/app/z.py --env APPNAME=z.py --env PORT=1111 -p 8000:1111  local/py_faas:0.0.1

FROM alpine

ARG http_p

ENV https_proxy $http_p
ENV http_proxy $http_p

ENV PORT 12345
ENV APPNAME ''

RUN apk --no-cache add \
            python \
            python-dev \
            py-pip \
            build-base

RUN pip install flask

WORKDIR /app

ADD display_output.py display_output.py
ADD default.py default.py

ENV http_proxy ''
ENV https_proxy ''

ENTRYPOINT ./display_output.py