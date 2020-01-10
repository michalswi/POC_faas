package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	logger.Printf("Starting server on port %s", servicePort)

	rserver := http.NewServeMux()

	rserver.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("Request dump: %v", r)
		stdout, stderr := exec.Command(workDir + appName).Output()
		if stderr != nil {
			log.Println("[-] Error in r.FormFile ", stderr)
			fmt.Fprintf(w, "%s", stderr)
		}
		fmt.Fprintf(w, "%s", stdout)
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%v", servicePort), rserver); err != nil {
		logger.Fatalf("Could not listen on %s: %v\n", servicePort, err)
	}

	logger.Printf("Server stopped")
}
