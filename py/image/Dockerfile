# PY: display output

# docker build -t local/py_faas:0.0.1 .
# docker run -it -v $(realpath .)/z.py:/app/z.py --env APPNAME=z.py --env PORT=1111 -p 8000:1111  local/py_faas:0.0.1

FROM alpine

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

ENTRYPOINT ./display_output.py