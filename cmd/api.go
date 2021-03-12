package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"tcw.im/tdi/tool"

	"github.com/spf13/cobra"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "运行API服务",
	Run: func(cmd *cobra.Command, args []string) {
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

		http.HandleFunc("/", Router)
		listen := fmt.Sprintf("%s:%d", host, port)
		log.Println("HTTP listen on " + listen)
		log.Fatal(http.ListenAndServe(listen, nil))
	},
}

func init() {
	apiCmd.Flags().SortFlags = false
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringVarP(&host, "host", "", "0.0.0.0", "Api监听地址")
	apiCmd.Flags().UintVarP(&port, "port", "", 13145, "Api监听端口")
	apiCmd.Flags().StringVarP(&token, "token", "t", "", "password to verify identity")
	apiCmd.Flags().StringVarP(&status, "status", "s", "ready", "set this service status: ready or tardy")
	apiCmd.Flags().StringVarP(&alert, "alert", "a", "ready", "set alarm mailbox")
}

func Router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Println(path)
	if strings.HasPrefix(path, "/ping") {
		pingView(w, r)
	}
}

func pingView(w http.ResponseWriter, r *http.Request) {
	info := make(map[string]interface{})
	memp, err := tool.MemRate()
	if err != nil {
		errView500(w, err)
	}
	diskp, err := tool.DiskRate(dir)
	if err != nil {
		errView500(w, err)
	}
	load5, err := tool.LoadStat()
	if err != nil {
		errView500(w, err)
	}
	info["code"] = 0
	info["version"] = version
	info["status"] = status
	info["memRate"] = memp
	info["diskRate"] = diskp
	info["loadFive"] = load5
	info["timestamp"] = time.Now().Unix()
	info["email"] = alert
	info["lang"] = runtime.Version()

	data, err := json.Marshal(info)
	if err != nil {
		errView500(w, err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func errView500(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
