package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
)

/*
// TODO
- recursive directory monitoring, nie dodany jeszcze do `fsnotify`

- POC:
- go cache
- boltdb

*/

const (
	upFolder    = "uploadsGO"
	ServicePort = ":5000"
	apiVersion  = "/api/v1"
	HostPort    = "8000"
	DockerPort  = "1111"
)

var upDir string
var goBinFile string
var getEventName string
var dockerIDvar string

func handleRequests(wg *sync.WaitGroup) {
	// myRouter := mux.NewRouter()
	// myRouter.HandleFunc("/emp", returnAllEmployees).Methods("GET")
	// myRouter.Path("/api/v1/getup/{folder}").Methods("GET").HandlerFunc(GetFiles)
	// myRouter.Path("/api/v1/up").Methods("POST").HandlerFunc(UploadFile)

	r := mux.NewRouter()

	// PathPrefix, na poziomie kodu nie trzeba podawac pelnej sciezki, w 'curl' tak
	myRouter := r.PathPrefix(apiVersion).Subrouter()
	// myRouter.StrictSlash(true)

	fmt.Printf("%+v\n", myRouter)
	// &{NotFoundHandler:<nil> MethodNotAllowedHandler:<nil> parent:0xc420112000
	// routes:[] namedRoutes:map[] strictSlash:false skipClean:false KeepContext:false
	// useEncodedPath:false}

	myRouter.Path("/").HandlerFunc(TestRoot)
	myRouter.Path("/getup/{folder}").Methods("GET").HandlerFunc(GetFiles)
	myRouter.Path("/up").Methods("POST").HandlerFunc(UploadFile)
	myRouter.Path("/stop/{id}").Methods("GET").HandlerFunc(stopDocker)

	fmt.Println("Start..")
	log.Fatal(http.ListenAndServe(ServicePort, myRouter))
	defer wg.Done()
}

func runDocker() {
	// https://docs.docker.com/develop/sdk/examples/#run-a-container
	// https://github.com/moby/moby/blob/master/api/types/container/config.go

	ctx := context.Background()
	// take the latest api version ($ docker version)
	// cli, err := client.NewEnvClient()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		panic(err)
	}

	imageName := "local/go_faas:0.0.1"
	// pull image
	// out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// io.Copy(os.Stdout, out)

	portToExpose := fmt.Sprintf("%s/tcp", DockerPort)
	// portHost := fmt.Sprintf("%s/tcp", HostPort)
	reqEnvs1 := fmt.Sprintf("PORT=%s", DockerPort)
	reqEnvs2 := fmt.Sprintf("APPNAME=%s", goBinFile)

	// go get github.com/docker/go-connections/nat
	// broken: mv ~/go/src/github.com/docker/docker/vendor/github.com/docker/go-connections/nat /tmp
	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: imageName,
			Env:   []string{reqEnvs1, reqEnvs2},
			ExposedPorts: nat.PortSet{
				nat.Port(portToExpose): {},
			},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: upDir + "/" + upFolder + "/" + goBinFile,
					Target: "/" + goBinFile,
				},
			},
			PortBindings: nat.PortMap{
				nat.Port(portToExpose): []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: HostPort},
				},
			},
		},
		nil,
		"",
	)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	dockerIDvar = resp.ID
	uploadResponse := fmt.Sprintf("[+] Docker ID: %s\n", dockerIDvar)
	fmt.Println(uploadResponse)
	return
}

func makeMainDirectory() {
	// by default, os.ModePerm = 0777
	// if err := os.MkdirAll("/root/"+upFolder, os.ModePerm); err != nil {
	if err := os.MkdirAll(upDir+"/"+upFolder, os.ModePerm); err != nil {
		log.Println("[-] Unable to create the directory. - " + err.Error())
		os.Exit(1)
		// panic("[-] Unable to create the directory. - " + err.Error())
	}
}

func TestRoot(w http.ResponseWriter, r *http.Request) {
	// curl -X GET localhost:5000/api/v1/ | jq

	// wyswietli w konsoli, ale po stronie serwera
	// fmt.Println("TestTest serwer")
	// wyswietli w konsoli, ale po stronie klienta
	// fmt.Fprintf(w, "TestTest klient")

	// wyswietli na stronie string
	// io.WriteString(w, "TestTest")

	// wyswietli na stronie json
	w.Header().Set("Content-Type	", "application/json; charset=utf-8")
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, `{"version":{"number":"0.0.1"}}`)
}

func stopDocker(w http.ResponseWriter, r *http.Request) {
	// curl -v localhost:5000/api/v1/stop/<id>

	params := mux.Vars(r)
	dockerID := params["id"]

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStop(ctx, dockerID, nil); err != nil {
		log.Println("[-]", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", err)
		return

	} else {
		log.Println("[+] Docker container stopped, id: " + dockerID)
		uploadResponse := fmt.Sprintf("[+] Docker container stopped, id: %s\n", dockerID)
		fmt.Fprintf(w, uploadResponse)
	}
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	// curl -X GET localhost:5000/api/v1/getup/uploadsGO | jq

	// https://gist.github.com/mattes/d13e273314c3b3ade33f
	// upFolder := "uploadsGO"

	// upDir, err := os.Getwd()
	// if err != nil {
	// 	log.Println("[-] Unable to get the realpath. - " + err.Error())
	// }

	var keys []string

	params := mux.Vars(r)

	if params["folder"] == upFolder {
		// list directory items
		files, err := ioutil.ReadDir(upDir + "/" + upFolder)
		if err != nil {
			// log.Fatal(err)
			log.Println("[-] Unable to get the directory. - " + err.Error())
		}
		for _, f := range files {
			// fmt.Println(f.Name())
			keys = append(keys, f.Name())
		}
		fmt.Println(keys)
		// return list, w - ResponseWriter
		json.NewEncoder(w).Encode(keys)
	} else {
		log.Println("[-] Missing or wrong directory.")
	}

}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	// curl -X POST -F 'file=@x.txt' localhost:5000/api/v1/up

	// TODO:
	// logging to a file using log. instead of fmt.
	// https://stackoverflow.com/questions/19965795/go-golang-write-log-to-file
	// https://www.dotnetperls.com/fprint-go

	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	// fmt.Println(io.Copy(ioutil.Discard, r.Body))

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("[-] Error in r.FormFile ", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{'error': %s}\n", err)
		return
	}
	defer file.Close()

	t := time.Now()
	// upFolder := "/uploadsGO"
	// upDir := "/tmp/uploadsGO"
	// upDir, err := os.Getwd()
	// if err != nil {
	// 	log.Println("[-] Unable to get the realpath. - " + err.Error())
	// }

	// directory creation moved to makeMainDirectory()
	// by default, os.ModePerm = 0777
	// if err := os.MkdirAll(upDir+"/"+upFolder, os.ModePerm); err != nil {
	// 	log.Println("[-] Unable to create the directory. - " + err.Error())
	// 	// panic("[-] Unable to create the directory. - " + err.Error())
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }

	// out, err := os.Create(filepath.Join(upDir, "uploaded-"+header.Filename))
	out, err := os.Create(filepath.Join(upDir+"/"+upFolder, t.Format("20060102150405-")+header.Filename))
	if err != nil {
		log.Println("[-] Unable to create the file for writing. Check your write access privilege.", err)
		// fmt.Fprintf(w, "[-] Unable to create the file for writing. Check your write access privilege.", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	// change permission to uploaded file(if not error when executing binary)
	out.Chmod(0777)
	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("[-] Error copying file.", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log.Println("[+] File uploaded successfully: uploaded-", header.Filename)
	uploadFileFormat := t.Format("20060102150405-") + header.Filename
	log.Println("[+] File uploaded successfully:", uploadFileFormat)
	uploadResponse := fmt.Sprintf("[+] File uploaded successfully: %s\n", uploadFileFormat)
	fmt.Fprintf(w, uploadResponse)

	// VERIFY IF GO BINARY
	// https://www.socketloop.com/tutorials/how-to-tell-if-a-binary-executable-file-or-web-application-is-built-with-golang
	// cat <binary> | grep Go
	// strings -20 <binary> | grep Go
	stdout, stderr := exec.Command("grep", "Go", getEventName).Output()
	if stderr != nil {
		log.Println("[-] Some internal error", stderr)
		return
	}

	if strings.Contains(string(stdout), "matches") {
		goBinTemp := strings.Split(getEventName, "/")
		// fmt.Printf("%+v\n", goBinTemp[len(goBinTemp)-1])
		goBinFile = goBinTemp[len(goBinTemp)-1]
		fmt.Printf("[+] Run docker..\n")
		fmt.Fprintf(w, "[+] Run docker..\n")
		//runDocker container
		runDocker()
		uploadResponse := fmt.Sprintf("[+] Docker ID: %s\n", dockerIDvar)
		fmt.Fprintf(w, uploadResponse)
	} else {
		log.Println("[-] It's not GO binary: " + goBinFile)
		fmt.Fprintf(w, "[-] It's not GO binary: "+goBinFile)
	}
}

func FileWatcher(wg *sync.WaitGroup) {
	// https://github.com/fsnotify/fsnotify/blob/master/example_test.go

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)

				if event.Op&fsnotify.Create == fsnotify.Create {
					getEventName = event.Name
				}

				// if event.Op&fsnotify.Write == fsnotify.Write {
				// 	log.Println("modified file:", event.Name)
				// }

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(upDir + "/" + upFolder)
	if err != nil {
		log.Fatal(err)
	}
	<-done
	defer wg.Done()
}

func main() {
	// f, _ := os.Create("/tmp/goserver.log")
	// w := bufio.NewWriter(f)
	fmt.Println("Rest API v1")

	var err error
	upDir, err = os.Getwd()
	if err != nil {
		log.Println("[-] Unable to get the realpath. - " + err.Error())
	}

	// handleRequests()
	makeMainDirectory()

	var wg sync.WaitGroup
	wg.Add(2)
	go handleRequests(&wg)
	go FileWatcher(&wg)
	wg.Wait()
	fmt.Println("DONE!")
}