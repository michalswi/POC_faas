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
- recursive directory monitoring, haven't beend added to `fsnotify` yet
https://github.com/qor/redirect_back
var RedirectBack = redirect_back.New(&redirect_back.Config{
  AllowedExtensions: []string{".txt", ""}
})

- volumes
https://docs.aws.amazon.com/sdk-for-go/api/service/s3/
https://github.com/aws/aws-sdk-go
*/

const (
	upFolder    = "uploadsGO"
	servicePort = ":5000"
	apiVersion  = "/api/v1"
	hostPort    = "8000"
	dockerPort  = "1111"
)

type Dinfo struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

var upDir string
var goBinFile string
var getEventName string
var dockerIDvar string

func handleRequests(wg *sync.WaitGroup) {
	logger := log.New(os.Stdout, "faas ", log.LstdFlags|log.Lshortfile)
	logger.Println("Server is starting...")

	r := mux.NewRouter()
	myRouter := r.PathPrefix(apiVersion).Subrouter()

	myRouter.Path("/").HandlerFunc(testRoot)
	myRouter.Path("/getup/{folder}").Methods("GET").HandlerFunc(getFiles)
	myRouter.Path("/up").Methods("POST").HandlerFunc(uploadFile)
	myRouter.Path("/stop/{id}").Methods("GET").HandlerFunc(stopDocker)
	myRouter.Path("/dockers").Methods("GET").HandlerFunc(getRunningDockers)
	myRouter.Path("/getout").Methods("GET").HandlerFunc(getOutput)

	err := http.ListenAndServe(servicePort, myRouter)
	if err != nil {
		logger.Fatalf("Start message error: %v", err)
	}
	defer wg.Done()
}

// curl localhost:5000/api/v1/getout
func getOutput(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("http://localhost:%s", hostPort)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("[-] Status code error: %d %s", res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Fprintf(w, string(body))
}

func runDocker() {
	ctx := context.Background()
	// take the latest api version ($ docker version)
	// cli, err := client.NewEnvClient()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		log.Fatalf("Unable to initialize a new API client. - %v\n", err)
	}

	imageName := "local/go_faas:0.0.1"
	// pull image
	// out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// io.Copy(os.Stdout, out)

	portToExpose := fmt.Sprintf("%s/tcp", dockerPort)
	// portHost := fmt.Sprintf("%s/tcp", hostPort)
	reqEnvs1 := fmt.Sprintf("PORT=%s", dockerPort)
	reqEnvs2 := fmt.Sprintf("APPNAME=%s", goBinFile)

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
					{HostIP: "0.0.0.0", HostPort: hostPort},
				},
			},
		},
		nil,
		"",
	)
	if err != nil {
		log.Fatalf("Unable to create container. - %v\n", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatalf("Unable to start a container. - %v\n", err)
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
		log.Fatalf("[-] Unable to create the directory. - %v\n", err)
	}
}

// curl -XGET localhost:5000/api/v1/ | jq
func testRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.WriteString(w, `{"version":{"number":"0.0.1"}}`)
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.Write([]byte("Hello in testRoot"))

}

// curl localhost:5000/api/v1/dockers
func getRunningDockers(w http.ResponseWriter, r *http.Request) {
	dckrs := make(map[string]Dinfo)
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		log.Fatalf("Running docker client: %v", err)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Running docker container list: %v", err)
	}
	for _, container := range containers {
		// fmt.Println(container.Names) // -> [/naughty_swirles]
		// fmt.Println(container.State)
		// fmt.Println(container.ID)
		// fmt.Println(container.Ports) // -> [{0.0.0.0 80 80 tcp}]
		dckrs[container.ID] = Dinfo{Name: strings.Join(container.Names, ""), State: container.State}
	}
	log.Println("[+] Get running Dockers")
	log.Println(dckrs)
	json.NewEncoder(w).Encode(dckrs)
}

// curl -v localhost:5000/api/v1/stop/<id>
func stopDocker(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dockerID := params["id"]

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		log.Fatalf("Unable to initialize: %v", err)
	}

	if err := cli.ContainerStop(ctx, dockerID, nil); err != nil {
		log.Println("[-]", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", err)
		return

	} else {
		log.Println("[+] Docker container stopped, id: " + dockerID)
		uploadResponse := fmt.Sprintf("[+] Docker container stopped, id: %s\n", dockerID)
		dockerIDvar = ""
		fmt.Fprintf(w, uploadResponse)
	}
}

// curl -X GET localhost:5000/api/v1/getup/uploadsGO | jq
func getFiles(w http.ResponseWriter, r *http.Request) {
	var keys []string
	params := mux.Vars(r)
	if params["folder"] == upFolder {
		// list directory items
		files, err := ioutil.ReadDir(upDir + "/" + upFolder)
		if err != nil {
			log.Printf("[-] Unable to get the directory. - \n" + err.Error())
		}
		for _, f := range files {
			keys = append(keys, f.Name())
		}
		fmt.Println(keys)
		json.NewEncoder(w).Encode(keys)
	} else {
		log.Printf("[-] Missing or wrong directory.\n")
	}
}

// curl -X POST -F 'file=@x.txt' localhost:5000/api/v1/up
func uploadFile(w http.ResponseWriter, r *http.Request) {
	// maximum allowed payload to 5 megabytes (file size)
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

	out, err := os.Create(filepath.Join(upDir+"/"+upFolder, t.Format("20060102150405-")+header.Filename))
	if err != nil {
		log.Printf("[-] Unable to create the file for writing. Check your write access privilege.\n", err)
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

	uploadFileFormat := t.Format("20060102150405-") + header.Filename
	log.Println("[+] File uploaded successfully:", uploadFileFormat)
	uploadResponse := fmt.Sprintf("[+] File uploaded successfully: %s\n", uploadFileFormat)
	fmt.Fprintf(w, uploadResponse)

	stdout, stderr := exec.Command("grep", "Go", getEventName).Output()
	if stderr != nil {
		log.Println("[-] Verify your uploaded file", stderr)
		fmt.Fprintf(w, "[-] Verify your uploaded file\n")
		return
	}

	if dockerIDvar == "" {
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
			ret_var := fmt.Sprintf("[-] It's not GO binary: %s" + goBinFile)
			log.Println(ret_var)
			fmt.Fprintf(w, ret_var)
		}
	} else {
		ret_var := fmt.Sprintf("[-] Stop the previous docker container, id: %s\n", dockerIDvar)
		log.Println(ret_var)
		fmt.Fprintf(w, ret_var)
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
