/*
   Copyright 2021 Hiroshi.tao

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"tcw.im/gtc"
)

const version = "0.2.5"

var (
	h bool
	v bool
	i bool

	noclean   bool // if true, do not delete download file, otherwise, auto delete
	cleanonce bool

	dir    string // download absolute path
	host   string
	port   uint
	token  string
	status string
	hour   uint // clean hour
)

const d = "downloads"

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(&h, "h", false, "show help")
	flag.BoolVar(&h, "help", false, "show help")

	flag.BoolVar(&v, "v", false, "show version and exit")
	flag.BoolVar(&v, "version", false, "show version and exit")

	flag.BoolVar(&i, "i", false, "")
	flag.BoolVar(&i, "info", false, "")

	flag.BoolVar(&cleanonce, "clean-once", false, "")
	flag.BoolVar(&noclean, "noclean", false, "")
	flag.UintVar(&hour, "hour", 12, "")

	flag.StringVar(&host, "host", "0.0.0.0", "")
	flag.UintVar(&port, "port", 13145, "")

	flag.StringVar(&dir, "d", "", "")
	flag.StringVar(&dir, "dir", "", "")

	flag.StringVar(&token, "t", "", "")
	flag.StringVar(&token, "token", "", "")

	flag.StringVar(&status, "s", "ready", "")
	flag.StringVar(&status, "status", "ready", "")

	flag.Usage = usage
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
	} else if v {
		fmt.Println(version)
	} else if i {
		checkDir := dir
		if checkDir == "" {
			checkDir = cwd()
		}
		dp, _ := diskRate(checkDir)
		mp, _ := memRate()
		fmt.Printf("Version:     %s\n", version)
		fmt.Printf("Go version:  %s\n", strings.TrimLeft(runtime.Version(), "go"))
		fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Disk Rate:   %.2f%%\n", dp)
		fmt.Printf("Memory Rate: %.2f%%\n", mp)
	} else {
		handle()
	}
}

func usage() {
	helpStr := `Usage: tdi [flags]

Doc to https://docs.saintic.com/tdi-go/
Git to https://github.com/staugur/tdi-go/

Flags:
  -h, --help            show this help message and exit
  -v, --version         show cli version and exit
  -i, --info            show version and system info
      --host            http listen host (default "0.0.0.0", env)
      --port            http listen port (default 13145, env)
      --hour            if clean, expiration time (default 12)
      --noclean         do not automatically clean up download files (env)
      --clean-once      manually clean up expired files (no run api)
  -d, --dir             download base directory (default "downloads", env)
  -t, --token           password to verify identity (required<random>, env)
  -s, --status          set service status: ready or tardy, (default "ready")
`
	fmt.Println(helpStr)
}

func handle() {
	if dir == "" {
		dir = os.Getenv("tdi_dir")
	}
	if dir == "" {
		dir = d
	}
	if !gtc.IsDir(dir) {
		gtc.CreateDir(dir)
	}
	if !path.IsAbs(dir) {
		dir = filepath.Join(cwd(), dir)
	}
	isRandomToken := false
	if token == "" {
		token = os.Getenv("tdi_token")
		if token == "" {
			token = genRandomString(8)
			isRandomToken = true
		}
	}
	if status == "" {
		status = os.Getenv("tdi_status")
	}
	if status != "tardy" {
		status = "ready"
	}
	if gtc.IsTrue(os.Getenv("tdi_noclean")) {
		noclean = true
	}
	envhost := os.Getenv("tdi_host")
	envport := os.Getenv("tdi_port")
	if envhost != "" {
		host = envhost
	}
	if envport != "" {
		envport, err := strconv.Atoi(envport)
		if err != nil {
			fmt.Println("Invalid environment tdi_port")
			return
		}
		port = uint(envport)
	}
	if hour <= 0 {
		fmt.Println("hour needs to be greater than 0")
		os.Exit(1)
	}
	// run clean download, only once, and exit
	fmt.Println("cleanonce", cleanonce)
	if cleanonce {
		cleanDownload(int(hour))
		os.Exit(0)
	}
	// start clean download task
	if !noclean {
		go func() {
			for {
				cleanDownload(int(hour))
				time.Sleep(time.Minute)
			}
		}()
	}
	// start api task
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = customHTTPErrorHandler
	e.GET("/ping", pingView)
	e.GET("/healthy", healthyView)
	e.POST("/download", downloadView)
	e.GET("/downloads/:filename", sendfileView)
	if isRandomToken {
		fmt.Println("the randomly generated token is: " + token)
	}
	address := fmt.Sprintf("%s:%d", host, port)
	fmt.Println("HTTP listen on " + address)
	e.Logger.Fatal(e.Start(address))
}
