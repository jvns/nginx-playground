package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config := os.Args[1]
	command := os.Args[2]
	result, err := run(config, command)
	resp := RunResponse{
		Result: result,
		Error:  toString(err),
	}
	jsonStr, _ := json.Marshal(resp)
	fmt.Println(string(jsonStr))
}

func toString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

type RunResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func run(nginxConfig string, command string) (string, error) {
	configFile := "/tmp/nginx_config"
	errorFile := "/tmp/nginx_errors"

	if err := os.WriteFile(configFile, []byte(nginxConfig), 0666); err != nil {
		return "", fmt.Errorf("Error creating %s: %s", configFile, err)
	}
	defer os.Remove(configFile)

	httpbin_cmd := exec.Command("go-httpbin", "-port", "7777")
	if err := httpbin_cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start go-httpbin: %s", err)
	}
	defer kill(httpbin_cmd)

	directives := fmt.Sprintf("daemon off; pid /tmp/%s.pid;", randSeq(16))
	nginx_cmd := exec.Command("nginx", "-c", configFile, "-e", errorFile, "-g", directives)
	nginx_cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	ch := make(chan error)
	go func() {
		ch <- nginx_cmd.Run()
	}()

	// Check for errors
	select {
	case <-ch:
		logs, _ := os.ReadFile(errorFile)
		return "", fmt.Errorf("nginx failed to start. Error logs:\n\n %s", string(logs))
	case <-time.After(100 * time.Millisecond):
		defer term(nginx_cmd)
		break
	}
	curlArgs := strings.Split(strings.TrimSpace(command), " ")

	if curlArgs[0] != "curl" && curlArgs[0] != "http" {
		return "", fmt.Errorf("command must start with 'curl' or 'http'")
	}

	output, err := exec.Command(curlArgs[0], curlArgs[1:]...).CombinedOutput()
	if err != nil {
		return string(output), err
	}

	logs, _ := os.ReadFile(errorFile)
	if strings.Contains(string(logs), "[error]") {
		err := fmt.Errorf("nginx error logs:\n\n %s", string(logs))
		return string(output), err
	}
	return string(output), nil
}

func kill(cmd *exec.Cmd) {
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
}

func term(cmd *exec.Cmd) {
	if cmd.Process != nil {
		cmd.Process.Signal(syscall.SIGTERM)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
