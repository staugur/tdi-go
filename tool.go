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
// common tools

package main

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"pkg.tcw.im/gtc"
)

func cwd() string {
	pwd, _ := os.Getwd()
	return pwd
}

func nowTimestamp() int64 {
	return time.Now().Unix()
}

// diskRate returns the usage rate of the disk where the directory is located
func diskRate(volumePath string) (percent float64, err error) {
	obj, e := disk.Usage(volumePath)
	if e != nil {
		err = e
		return
	}
	return strconv.ParseFloat(fmt.Sprintf("%.2f", obj.UsedPercent), 64)
}

// memRate returns system memory usage
func memRate() (percent float64, err error) {
	obj, e := mem.VirtualMemory()
	if e != nil {
		err = e
		return
	}
	return strconv.ParseFloat(fmt.Sprintf("%.2f", obj.UsedPercent), 64)
}

// loadStat returns load value in 5 minutes(float)
func loadStat() (loadavg5 float64, err error) {
	obj, e := load.Avg()
	if e != nil {
		err = e
		return
	}
	loadavg5 = obj.Load5
	return
}

// makeTarFile compress all files in a directory.
// Automatically delete after compression.
func makeTarFile(tarFilename, tarPath string, exclude []string) (err error) {
	if !strings.HasSuffix(tarFilename, ".tar") || !gtc.IsDir(tarPath) {
		return errors.New("make tar: invalid param")
	}
	fw, err := os.Create(tarFilename)
	if err != nil {
		return
	}
	defer fw.Close()

	// create Tar.Writer structure
	tw := tar.NewWriter(fw)
	defer tw.Close()

	// Recursively process all files in the directory
	return filepath.Walk(tarPath, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// if the filename suffix (such as .gz .xxx) is in the exclusion list,
		// do not compress
		if gtc.StrInSlice(path.Ext(fileName), exclude) {
			return nil
		}
		// if not is a standard file, do not process it, such as: directory
		if !fi.Mode().IsRegular() {
			return nil
		}
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		// write file information
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		fr, err := os.Open(fileName)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, fr)
		if err != nil {
			return err
		}
		fr.Close()
		os.Remove(fileName)
		return nil
	})
}

// formatSize format byte size as kilobytes, megabytes, gigabytes
func formatSize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func SHA1(text string) string {
	if text == "" {
		return ""
	}
	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func seriaName(filename string) string {
	td := os.TempDir()
	if !gtc.IsDir(td) {
		td = dir
	}
	return filepath.Join(td, fmt.Sprintf(".%s.dat", filename))
}

func serialize(data interface{}, filename string) error {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}
	return os.WriteFile(seriaName(filename), buffer.Bytes(), 0600)
}

func deserialize(data interface{}, filename string) error {
	raw, err := os.ReadFile(seriaName(filename))
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buffer)
	return dec.Decode(data)
}

func rmserialize(filename string) error {
	return os.Remove(seriaName(filename))
}

func genRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
