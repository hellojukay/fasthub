package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Meta github api response struct
type Meta struct {
	Web []string `json:"web"`
	Git []string `json:"git"`
}

func githubIPS(api string) ([]string, error) {
	var client = http.DefaultClient
	res, err := client.Get(api)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var meta Meta
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &meta)
	var ips = append(meta.Git, meta.Web...)
	var result []string
	for _, ip := range ips {
		result = append(result, strings.Split(ip, "/")[0])
	}
	return result, nil
}

// check ssh connection time
func checkSSH(ip string) (int, error) {
	var startTime = time.Now().Nanosecond()
	_, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", ip), time.Duration(10)*time.Second)
	if err != nil {
		return 0, err
	}
	var endTime = time.Now().Nanosecond()
	return endTime - startTime, nil
}

//https://api.github.com/meta
func main() {
	var waitGroup sync.WaitGroup
	ips, err := githubIPS("https://api.github.com/meta")
	if err != nil {
		fmt.Printf("request github error %s\n", err)
		os.Exit(1)
	}
	for _, ip := range ips {
		waitGroup.Add(1)
		go func(ip string) {
			t, err := checkSSH(ip)
			if err != nil {
				fmt.Printf("%-20s %10s", ip, err)
			} else {
				fmt.Printf("%-20s %10d", ip, t)
			}
		}(ip)
	}
	waitGroup.Wait()
}
