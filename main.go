package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"tcw.im/ufc"
)

const version = "0.1.0"

var (
	h bool
	v bool
	i bool

	noclean bool // if true, do not delete download file, otherwise, auto delete

	dir    string // download absolute path
	host   string
	port   uint
	rawurl string // redis connect url
	token  string
	status string
	alert  string // alert email

	rc *redis.Client
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

	flag.StringVar(&host, "host", "0.0.0.0", "")
	flag.UintVar(&port, "port", 13145, "")

	flag.StringVar(&dir, "d", "", "")
	flag.StringVar(&dir, "dir", "", "")

	flag.StringVar(&rawurl, "r", "", "")
	flag.StringVar(&rawurl, "redis", "", "")

	flag.StringVar(&token, "t", "", "")
	flag.StringVar(&token, "token", "", "")

	flag.StringVar(&status, "s", "ready", "")
	flag.StringVar(&status, "status", "ready", "")

	flag.StringVar(&alert, "a", "", "")
	flag.StringVar(&alert, "alert", "", "")

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
		fmt.Printf("Goroutine:   %d\n", runtime.NumGoroutine())
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
      --host            http listen host (default "0.0.0.0", env)
      --port            http listen port (default 13145, env)
  -d, --dir             download base directory (required)
  -r, --redis           redis url, DSN-Style (required, env)
                        format: redis://[:<password>@]host[:port/db]
  -t, --token           password to verify identity (required, env)
  -s, --status          set this service status: ready or tardy, (default ready)
  -a, --alert           set alarm mailbox (env)
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
	if rawurl == "" {
		rawurl = os.Getenv("tdi_redis")
		if rawurl == "" {
			fmt.Println("invalid environment tdi_redis")
			os.Exit(128)
		}
	}
	if token == "" {
		token = os.Getenv("tdi_token")
		if token == "" {
			fmt.Println("invalid environment tdi_token")
			os.Exit(129)
		}
	}
	if alert == "" {
		alert = os.Getenv("tdi_alert")
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

	// init redis connect client
	opt, err := redis.ParseURL(rawurl)
	if err != nil {
		fmt.Println(err)
		return
	}
	rc = redis.NewClient(opt)

	if !noclean {
		go func() {
			for {
				cleanDownload(12)
				time.Sleep(time.Minute)
			}
		}()
	}

	// view.go
	http.HandleFunc("/", router)
	listen := fmt.Sprintf("%s:%d", host, port)
	log.Println("HTTP listen on " + listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
