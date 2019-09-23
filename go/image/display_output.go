package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

var servicePort = getEnv("PORT", "8000")
var appName = getEnv("APPNAME", "default")
var workDir = "./"

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func main() {
	logger := log.New(os.Stdout, "executor ", log.LstdFlags|log.Lshortfile)
	logger.Println("Server is starting...")

	myRouter := mux.NewRouter()
	myRouter.Path("/").Methods("GET").HandlerFunc(DispOut)

	// log.Fatal(http.ListenAndServe(":"+strconv.Itoa(ServicePort), myRouter))
	logger.Fatal(http.ListenAndServe(":"+servicePort, myRouter))
}

func DispOut(w http.ResponseWriter, r *http.Request) {
	stdout, stderr := exec.Command(workDir + appName).Output()
	if stderr != nil {
		log.Println("[-] Error in r.FormFile ", stderr)
		fmt.Fprintf(w, "%s", stderr)
	}
	fmt.Fprintf(w, "%s", stdout)
}
