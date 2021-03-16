// download in web & clean download in cli

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"tcw.im/ufc"
)

func downloadBoard(data *download) {
	log.Printf("download start for %s in %s\n", data.Uifn, cwd())

	pins := data.downloads
	maxs := int(data.MAXBoardNumber)
	if len(pins) > maxs {
		pins = pins[:maxs]
	}
	var readme strings.Builder
	var allowDown bool = true
	dp, err := diskRate(dir)
	if err != nil {
		allowDown = false
		readme.WriteString(err.Error())
	}
	if dp > data.DiskLimit {
		allowDown = false
		readme.WriteString("disk usage is too high")
	}

	os.Chdir(dir)
	err = ufc.CreateDir(data.BoardId)
	if err != nil {
		allowDown = false
		readme.WriteString("create board directory failed")
	}
	// Root directory of current and subsequent coroutines
	os.Chdir(data.BoardId)

	// if allowDown is false, abort the program
	if !allowDown {
		ioutil.WriteFile("README.txt", []byte(readme.String()), 0755)
		return
	}

	// construct the request header
	var ref string
	if data.Site == 1 {
		ref = fmt.Sprintf("https://huaban.com/boards/%s", data.BoardId)
	} else {
		ref = fmt.Sprintf("https://www.duitang.com/album/?id=%s", data.BoardId)
	}
	headers := make(map[string]string)
	headers["Referer"] = ref
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"

	// start to download
	nt := nowTimestamp()
	var wg sync.WaitGroup
	for _, p := range pins {
		wg.Add(1)
		go func(p pin) {
			defer wg.Done()
			if ufc.IsFile(p.Name) {
				return
			}
			dp, _ := diskRate(dir)
			if dp > data.DiskLimit {
				readme.WriteString("disk usage is too high")
				return
			}
			resp, err := httpGet(p.URL, headers)
			if err != nil {
				readme.WriteString(err.Error())
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				readme.WriteString(err.Error())
				return
			}
			ioutil.WriteFile(p.Name, body, 0755)
			time.Sleep(50 * time.Millisecond)
		}(p)
	}
	wg.Wait()
	if readme.Len() > 0 {
		ioutil.WriteFile("README.txt", []byte(readme.String()), 0755)
	}
	os.Chdir(dir)
	log.Println("downloading end, make tar")

	dtime := nowTimestamp() - nt
	exclude := []string{".zip", ".lock", ".tar"}
	err = makeTarFile(data.Uifn, data.BoardId, exclude)
	if err != nil {
		log.Println(err)
		return
	}
	ui, err := os.Stat(data.Uifn)
	if err != nil {
		log.Println(err)
		return
	}
	size := formatSize(ui.Size())
	os.Remove(data.BoardId)
	body := make(map[string]string)
	body["uifn"] = data.Uifn
	body["uifnKey"] = data.UifnKey
	body["size"] = size
	body["dtime"] = fmt.Sprintf("%d", dtime)
	resp, err := httpPost(data.CallbackURL+"?Action=FIRST_STATUS", body)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("download over, success")
}

// perform a cleanup
func cleanDownload(hours int) {
	log.Println("clean download")
	dfs, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, f := range dfs {
		if !f.Mode().IsRegular() {
			continue
		}
		n := f.Name()
		if path.Ext(n) != ".tar" {
			continue
		}
		ns := strings.Split(strings.TrimSuffix(n, path.Ext(n)), "_")
		if len(ns) < 2 {
			continue
		}
		aid := ns[0]
		mst, err := strconv.Atoi(ns[1])
		if err != nil {
			continue
		}
		if aid != "hb" {
			continue
		}
		// real process
		ctime := mst / 1000
		fctime := f.ModTime()
		log.Println(fctime)
		nt := nowTimestamp()
		if (ctime + 60*60*hours) <= int(nt) {
			// expired, clean and report
			val, err := rc.Get(context.Background(), n).Result()
			if err != nil {
				continue
			}

			data := &download{}
			err = json.Unmarshal([]byte(val), data)
			if err != nil {
				continue
			}
			body := make(map[string]string)
			body["uifn"] = data.Uifn

			resp, err := httpPost(data.CallbackURL+"?Action=SECOND_STATUS", body)
			if err != nil {
				log.Println(err)
				continue
			}
			text, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}
			os.Remove(n)
			log.Printf("Update expired status for %s, resp is %s", n, string(text))
		}
	}
}
