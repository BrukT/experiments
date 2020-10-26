package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

var KubectlPath string = "/usr/local/bin/kubectl"
var deviceName string = "coap-time-client"
var configPath string = "/home/ubuntu/.kube/config"

func main() {
	var buf = new(bytes.Buffer)
	cmdKube := &exec.Cmd{
		Path:   KubectlPath,
		Args:   []string{KubectlPath, "--kubeconfig=" + configPath, "get", "device", deviceName, "-o", "json", "-w"},
		Stdout: buf,
		Stderr: os.Stdout,
	}

	go func() {
		if err := cmdKube.Run(); err != nil {
			log.Fatal("Getting the state", err)
		}
	}()
	for {
		fmt.Print(buf.String())
		time.Sleep(time.Second)
		// dat := make(map[string]interface{})
	}
	// if err := json.Unmarshal([]byte(buf.String()), &dat); err != nil {
	// 	panic(err)
	// }
	// fmt.Println(dat)
}
