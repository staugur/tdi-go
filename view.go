package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Println(path)
	if strings.HasPrefix(path, "/ping") {
		pingView(w, r)
	}
}

func pingView(w http.ResponseWriter, r *http.Request) {
	info := make(map[string]interface{})
	memp, err := memRate()
	if err != nil {
		errView500(w, err)
	}
	diskp, err := diskRate(dir)
	if err != nil {
		errView500(w, err)
	}
	load5, err := loadStat()
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
