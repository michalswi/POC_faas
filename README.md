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

#### Test 'image/display_output.go'

**APPNAME** should be available before 'display_output.go' run.  
**APP** should be a binary file.  

```sh
$ ./run.sh

#
$ docker run --rm -p 8000:8000 local/go_faas:0.0.1
executor 2020/01/10 08:00:53 display_output.go:25: Starting server on port 8000

$ curl localhost:8000
Something went wrong, your app/script was not uploaded..

#
$ docker run --rm -v $(realpath image/bingo/test):/test -e APPNAME=test -p 8000:8000 local/go_faas:0.0.1
executor 2020/01/10 08:00:53 display_output.go:25: Starting server on port 8000

$ curl localhost:8000
Test Test test..

$ ./delete.sh
```

#### Test 'up_file_local.go'
```sh

# step 1
[tty1]
$ ./run.sh
$ ./up_file_local


[tty2] 
<check APIs below>


# step 2
[tty1]
$ ./delete.sh
```

#### APIs
- Test root  
`$ curl localhost:5000/api/v1/ | jq`

- Upload file  
`$ curl -X POST -F 'file=@image/bingo/test' localhost:5000/api/v1/up`

- Get running dockers  
`$ curl localhost:5000/api/v1/dockers`

- Check uploaded files  
`$ curl localhost:5000/api/v1/getup/uploadsGO | jq`

- Get output from running container  
`$ curl localhost:5000/api/v1/getout`

- Stop docker  
`$ curl localhost:5000/api/v1/stop/<id>`