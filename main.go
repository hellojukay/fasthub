// https://github.com/hellojukay/fasthub
package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
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

const (
	// HTTP http connection  ip
	HTTP = "HTTP"
	// SSH ssh conection ip
	SSH = "SSH"
)

// Address git ip address
type Address struct {
	IP string
	// SSH or http
	TYPE string
}

func githubIPS(api string) ([]Address, error) {
	var client = http.DefaultClient
	res, err := client.Get(api)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var meta Meta
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &meta)
	var result []Address
	for _, ip := range meta.Web {
		var arr = strings.Split(ip, "/")
		if arr[1] != "32" {
			continue
		}
		result = append(result, Address{
			IP:   arr[0],
			TYPE: HTTP,
		})
	}
	for _, ip := range meta.Git {
		var arr = strings.Split(ip, "/")
		if arr[1] != "32" {
			continue
		}
		result = append(result, Address{
			IP:   arr[0],
			TYPE: SSH,
		})
	}
	return result, nil
}

// check ssh connection time
func checkSSH(ip string) (float64, error) {
	var startTime = time.Now()
	con, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", ip), time.Duration(10)*time.Second)
	if err != nil {
		return 0, errors.New("TIMEOUT")
	}
	con.SetDeadline(time.Now().Add(time.Second * 5))
	var buf = make([]byte, 1)
	con.Read(buf)
	var endTime = time.Now()
	var duration = endTime.Sub(startTime)
	return duration.Seconds(), nil
}

// check http connection time
func checkHTTP(ip string) (float64, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var startTime = time.Now()
	res, err := http.Get(fmt.Sprintf("http://%s/hellojukay", ip))
	if err != nil {
		return 0, errors.New("TIMEOUT")
	}
	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, errors.New("TIMEOUT")
	}
	return time.Now().Sub(startTime).Seconds(), nil
}

//https://api.github.com/meta
func main() {
	var waitGroup sync.WaitGroup
	addresses, err := githubIPS("https://api.github.com/meta")
	if err != nil {
		fmt.Printf("request github error %s\n", err)
		os.Exit(1)
	}
	for _, ip := range addresses {
		waitGroup.Add(1)
		go func(address Address) {
			var t float64
			var err error
			if address.TYPE == SSH {
				t, err = checkSSH(address.IP)
			} else {
				t, err = checkHTTP(address.IP)
			}
			if err != nil {
				fmt.Printf("%-20s%6s%8s\n", address.IP, err, address.TYPE)
			} else {
				fmt.Printf("%-20s%6s%8s\n", address.IP, fmt.Sprintf("%2.2fs", t), address.TYPE)
			}
			waitGroup.Done()
		}(ip)
	}
	waitGroup.Wait()
}
