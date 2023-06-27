package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:embed "healthcheck.json"
var healthcheck string

type RunRequest struct {
	NginxConfig string `json:"nginx_config"`
	Command     string `json:"command"`
}
type RunResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.Handle("/", wrapLogger(http.HandlerFunc(runHandler)))
	fmt.Println("Listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// make a request to localhost:8080 with `healthcheck` as the body
	// if it works, return 200
	// if it doesn't, return 500
	client := http.Client{}
	resp, err := client.Post("http://localhost:8080/", "application/json", strings.NewReader(healthcheck))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Println(resp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "*")
	if r.Method != "POST" {
		// OPTIONS request
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	var req RunRequest
	json.Unmarshal([]byte(body), &req)

	logdir, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(logdir)
	tmpdir, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	rootdir, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(rootdir)
	cachedir, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(cachedir)
	pwd, _ := os.Getwd()
	cmd := exec.Command("bwrap",
		"--ro-bind", "/bin", "/bin",
		"--ro-bind", "/etc", "/etc",
		"--ro-bind", "/lib", "/lib",
		"--ro-bind", "/opt", "/opt",
		"--ro-bind", "/usr", "/usr",
		"--ro-bind", pwd, "/app",
		"--ro-bind-try", "/lib32", "/lib32",
		"--ro-bind-try", "/lib64", "/lib64",
		"--unshare-net",
		"--unshare-pid",
		"--dev", "/dev",
		"--proc", "/proc",
		"--bind-try", "/var/lib/nginx/", "/var/lib/nginx/", // only on laptop
		"--dir", "/var/log/nginx",
		"--dir", "/tmp",
		"--dir", "/root",
		"--dir", "/var/cache/nginx",

		"/app/run_nginx", req.NginxConfig, req.Command,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error running command: ", err, string(output))
		return
	}
	var resp RunResponse
	err = json.Unmarshal(output, &resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("error unmarshalling: ", err, "'"+string(output)+"'")
		return
	}
	if resp.Error != "" {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(output)
}

func wrapLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		elapsed := time.Since(start)
		log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL.Path, elapsed)
	})
}
