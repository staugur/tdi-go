// web api view

package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"runtime"
	"time"
)

func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	w.Header().Set("Server", "tdi/go")
	w.Header().Set("Content-Type", "application/json")
	err := signatureRequired(r)
	if err != nil {
		errView(w, err)
		return
	}
	if path == "/ping" {
		if r.Method != "GET" {
			errView405(w)
			return
		}
		pingView(w, r)
	} else if path == "/download" {
		if r.Method != "POST" {
			errView405(w)
			return
		}
		downloadView(w, r)
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
	info["rqcount"] = -1
	info["rqfailed"] = -1
	info["goroutine"] = runtime.NumGoroutine()

	data, err := json.Marshal(info)
	if err != nil {
		errView500(w, err)
	}
	w.Write(data)
}

func downloadView(w http.ResponseWriter, r *http.Request) {
	data := &download{}
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		errView400(w)
		return
	}
	if data.UifnKey == "" || data.Uifn == "" || data.BoardId == "" ||
		data.BoardPins == "" || data.Ctime == 0 || data.Etime == 0 {
		errView(w, errors.New("invalid param"))
		return
	}

	pins := make([]pin, 0)
	json.Unmarshal([]byte(data.BoardPins), &pins)
	if len(pins) < 1 {
		errView(w, errors.New("empty download"))
		return
	}
	data.downloads = pins

	// write to redis
	dumps, err := json.Marshal(data)
	if err != nil {
		errView(w, err)
		return
	}
	rc.Set(context.Background(), data.Uifn, dumps, 7*24*time.Hour)

	go downloadBoard(data)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"code":0,"msg":"downloading"}`))
}
