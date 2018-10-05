package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

// APPNAME should be available before 'display_output.go' run
// APP should be a binary file

var ServicePort = getEnv("PORT", "8000")
var AppName = getEnv("APPNAME", "default")

// var AppName = getEnv("APPNAME", "default.go")

// getEnv get key environment variable if exist otherwise return defalutValue
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func handleRequests() {
	myRouter := mux.NewRouter()
	myRouter.Path("/").Methods("GET").HandlerFunc(DispOut)
	fmt.Println("Start..")
	// log.Fatal(http.ListenAndServe(":"+strconv.Itoa(ServicePort), myRouter))
	log.Fatal(http.ListenAndServe(":"+ServicePort, myRouter))
}

func DispOut(w http.ResponseWriter, r *http.Request) {

	// https://stackoverflow.com/questions/8875038/redirect-stdout-pipe-of-child-process-in-go
	// cmd := exec.Command("go", "run", "default.go")

	// cmd := "dmes"
	// stdout, stderr := exec.Command(cmd).Output()

	// uncomment var AppName
	// stdout, stderr := exec.Command("go", "run", AppName).Output()

	stdout, stderr := exec.Command("./" + AppName).Output()
	if stderr != nil {
		log.Println("[-] Error in r.FormFile ", stderr)
		fmt.Fprintf(w, "%s", stderr)
	}
	fmt.Fprintf(w, "%s", stdout)

	// cmd := exec.Command("dmesg")
	// f, err := pty.Start(cmd)
	// if err != nil {
	// 	// panic(err)
	// 	log.Println("[-] Error in r.FormFile ", err)
	// 	fmt.Fprintf(w, "{'error': %s}\n", err)
	// }
	// io.Copy(w, f)
}

func main() {
	handleRequests()
}
