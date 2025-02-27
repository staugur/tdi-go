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
// download in web & clean download in cli

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"pkg.tcw.im/gtc"
)

type strbuilder struct {
	b strings.Builder
}

func (s strbuilder) WriteS(text string) {
	s.b.WriteString(text + "\n")
}

func (s strbuilder) WriteE(err error) {
	s.b.WriteString(err.Error() + "\n")
}

func (s strbuilder) String() string {
	return s.b.String()
}

func (s strbuilder) Len() int {
	return s.b.Len()
}

func (s strbuilder) FlushReadme() error {
	return os.WriteFile("README.txt", []byte(s.String()), 0755)
}

func downloadBoard(data *download) {
	log.Printf("download start for %s in %s\n", data.Uifn, dir)

	pins := data.downloads
	maxs := int(data.MAXBoardNumber)
	readme := strbuilder{}
	var allowDown bool = true
	if len(pins) > maxs {
		pins = pins[:maxs]
	}

	var gs int // Divide pins into some number parts
	var maxLimit int = 20
	if len(pins) > 5000 {
		maxLimit = 50
	} else if len(pins) > 10000 {
		maxLimit = 100
	}
	if len(pins) > maxLimit {
		gs = len(pins) / maxLimit
	} else {
		gs = 1
	}

	dp, err := diskRate(dir)
	if err != nil {
		allowDown = false
		readme.WriteE(err)
	}
	if dp > data.DiskLimit {
		allowDown = false
		readme.WriteS("disk usage is too high")
	}

	err = os.Chdir(dir)
	if err != nil {
		log.Println(err.Error())
	}
	err = gtc.CreateDir(data.BoardId)
	if err != nil {
		log.Printf("create board directory failed: %s\n", err.Error())
		return
	}
	// root directory of current and subsequent coroutines
	os.Chdir(data.BoardId)

	// split download pins
	spins, err := splitPins(pins, gs)
	if err != nil {
		allowDown = false
		readme.WriteE(err)
	}

	// if allowDown is false, abort the program
	if !allowDown {
		log.Println("system judgment is not allowed to download")
		readme.FlushReadme()
		return
	}

	// start to download
	nt := nowTimestamp()
	// construct the request header
	var ref string
	if data.Site == 1 {
		ref = fmt.Sprintf("https://huaban.com/boards/%s", data.BoardId)
	} else {
		ref = fmt.Sprintf("https://www.duitang.com/album/?id=%s", data.BoardId)
	}
	headers := make(map[string]string)
	headers["Referer"] = ref
	headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0"

	var wg sync.WaitGroup
	for _, sp := range spins {
		wg.Add(1)
		// Download a set of pictures for each coroutine
		go func(sp []pin) {
			defer wg.Done()
			for _, p := range sp {
				func(p pin) {
					if gtc.IsFile(p.Name) {
						return
					}
					dp, _ := diskRate(dir)
					if dp > data.DiskLimit {
						readme.WriteS("disk usage is too high")
						return
					}
					var retry time.Duration = 1
					var resp *http.Response
					var err error
					for retry <= 3 {
						resp, err = httpGet(p.URL, headers, retry*10*time.Second)
						if err == nil {
							break
						}
						retry++
					}
					if err != nil {
						readme.WriteE(err)
						return
					}
					defer resp.Body.Close()
					pf, err := os.Create(p.Name)
					if err != nil {
						readme.WriteE(err)
						return
					}
					defer pf.Close()
					io.Copy(pf, resp.Body)
					time.Sleep(10 * time.Millisecond)
				}(p)
			}
		}(sp)
	}
	wg.Wait()
	if readme.Len() > 0 {
		log.Println("discover warning tips for Readme.txt")
		readme.FlushReadme()
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
	defer os.Remove(data.BoardId)
	size := formatSize(ui.Size())
	body := make(map[string]string)
	body["uifn"] = data.Uifn
	body["uifnKey"] = data.UifnKey
	body["size"] = size
	body["dtime"] = fmt.Sprintf("%d", dtime)
	log.Printf("Post data: %v\n", body)
	resp, err := httpPost(fmt.Sprintf("%s?Action=FIRST_STATUS", data.CallbackURL), body)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	text, err := io.ReadAll(resp.Body)
	if err == nil {
		log.Printf("Update download status for %s, resp is %s\n", data.Uifn, string(text))
	} else {
		log.Println("Read first post callback result failed")
	}

	log.Println("download over, successfully")
}

// perform a cleanup
func cleanDownload(hours int) {
	dfs, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, f := range dfs {
		if !f.Type().IsRegular() {
			continue
		}
		// n is Uifn
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
		// checked pass, enter the processing flow
		fi, fierr := f.Info()
		if fierr != nil {
			continue
		}
		ctime := mst / 1000
		fctime := fi.ModTime().Unix()
		ltime := 60 * 60 * hours
		nt := nowTimestamp()
		if (ctime+ltime) <= int(nt) && (fctime+int64(ltime)) <= nt {
			// expired, clean and report
			var data clean
			err := deserialize(&data, n)
			if err != nil {
				continue
			}

			body := make(map[string]string)
			body["uifn"] = data.Uifn
			resp, err := httpPost(fmt.Sprintf("%s?Action=SECOND_STATUS", data.CallbackURL), body)
			if err != nil {
				log.Println(err)
				continue
			}
			defer resp.Body.Close()
			text, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			rmserialize(n)
			os.Remove(filepath.Join(dir, n))
			log.Printf("Update expired status for %s, resp is %s", n, string(text))
		}
	}
}
