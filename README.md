### Python [todo]
```sh
# https://pypi.org/project/inotify/
$ pip install inotify

# https://pypi.org/project/docker/
$ pip install --user docker
```

### GO
```sh
$ go get github.com/kr/pty
$ go get github.com/gorilla/mux
$ go get github.com/fsnotify/fsnotify
$ go get github.com/docker/docker/client
$ go get github.com/docker/go-connections/nat
if broken: 
$ mv ~/go/src/github.com/docker/docker/vendor/github.com/docker/go-connections/nat /tmp
```

### How to

#### 'image/display_output.go'
**APPNAME** should be available before 'display_output.go' run.  
**APP** should be a binary file.  

```sh
$ ./run.sh

#
$ docker run -p 8000:8000 local/go_faas:0.0.1
$ curl localhost:8000
Something went wrong, your app/script was not uploaded..

#
$ docker run -v $(realpath image/bingo/test):/test -e APPNAME=test -p 8000:8000 local/go_faas:0.0.1
$ curl localhost:8000
Test Test test..

$ ./delete.sh
```

#### 'up_file_local.go'
```sh
[tty1]
$ ./run.sh
$ ./up_file_local
$ ./delete.sh

[tty2] 
<check APIs>
```

#### APIs
- Test root  
`$ curl localhost:5000/api/v1/ | jq`

- Upload file  
`$ curl -X POST -F 'file=@image/bingo/test' localhost:5000/api/v1/up`

- Get running dockers  
`$ curl localhost:5000/api/v1/dockers`

- Stop docker  
`$ curl localhost:5000/api/v1/stop/<id>`

- Check uploaded files  
`$ curl localhost:5000/api/v1/getup/uploadsGO | jq`

- Get output from running container  
`$ curl localhost:5000/api/v1/getout`

#### Examples
```sh
$ curl -X POST -F 'file=@examples/biny' localhost:5000/api/v1/up
{'error': http: request body too large}

$ curl -X POST -F 'file=@examples/file.txt' localhost:5000/api/v1/up
[+] File uploaded successfully: 20190924111139-file.txt
[-] Verify your uploaded file
```