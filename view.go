package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Println(path)
	w.Header().Set("Server", "tdi/go")
	w.Header().Set("Content-Type", "application/json")
	if strings.HasPrefix(path, "/ping") {
		pingView(w, r)
	} else {
		errView404(w)
	}
}

func pingView(w http.ResponseWriter, r *http.Request) {
	info := make(map[string]interface{})
	load5, err := loadStat()
	if err != nil {
		errView500(w, err)
		return
	}
	memp, err := memRate()
	if err != nil {
		errView500(w, err)
		return
	}
	diskp, err := diskRate(dir)
	if err != nil {
		errView500(w, err)
		return
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
	w.Write(data)
}

func errView500(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, `{"code":500,"msg":"%s"}`, err.Error())
}

func errView404(w http.ResponseWriter) {
	w.Write([]byte(`{"code":404,"msg":"not found page"}`))
}
