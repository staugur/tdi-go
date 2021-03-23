package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"tcw.im/ufc"
)

const version = "0.2.0"

var (
	h bool
	v bool
	i bool

	noclean bool // if true, do not delete download file, otherwise, auto delete

	dir    string // download absolute path
	host   string
	port   uint
	token  string
	status string
	hour   uint // clean hour
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(&h, "h", false, "show help")
	flag.BoolVar(&h, "help", false, "show help")

	flag.BoolVar(&v, "v", false, "show version and exit")
	flag.BoolVar(&v, "version", false, "show version and exit")

	flag.BoolVar(&i, "i", false, "")
	flag.BoolVar(&i, "info", false, "")

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
      --noclean         do not automatically clean up download files (env)
      --hour            if clean, expiration time (default 12)
      --host            http listen host (default "0.0.0.0", env)
      --port            http listen port (default 13145, env)
  -d, --dir             download base directory (required, env)
  -t, --token           password to verify identity (required, env)
  -s, --status          set this service status: ready or tardy, (default ready)
`
	fmt.Println(helpStr)
}

func handle() {
	if dir == "" {
		dir = os.Getenv("tdi_dir")
	}
	if dir == "" {
		fmt.Println("invalid dir")
		os.Exit(127)
	}
	if !ufc.IsDir(dir) {
		ufc.CreateDir(dir)
	}
	if !path.IsAbs(dir) {
		dir = filepath.Join(cwd(), dir)
	}
	if token == "" {
		token = os.Getenv("tdi_token")
		if token == "" {
			fmt.Println("invalid environment tdi_token")
			os.Exit(129)
		}
	}
	if status == "" {
		status = os.Getenv("tdi_status")
	}
	if status != "tardy" {
		status = "ready"
	}
	if ufc.IsTrue(os.Getenv("tdi_noclean")) {
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

	if !noclean {
		if hour <= 0 {
			hour = 12
		}
		go func() {
			for {
				cleanDownload(int(hour))
				time.Sleep(time.Minute)
			}
		}()
	}

	// view.go
	mime.AddExtensionType(".tar", "application/octet-stream")
	http.HandleFunc("/", router)
	http.Handle(
		"/downloads/",
		http.StripPrefix("/downloads", http.FileServer(http.Dir(dir))),
	)
	listen := fmt.Sprintf("%s:%d", host, port)
	log.Println("HTTP listen on " + listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
