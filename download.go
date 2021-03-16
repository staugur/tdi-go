// download in web & clean download in cli

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36"

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

func cleanDownload() {
	log.Println("clean download")
	for {
		log.Println("cleaning")
		time.Sleep(time.Minute)
	}
}
