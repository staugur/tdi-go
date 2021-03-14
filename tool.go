// common tools

package main

import (
	"archive/tar"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tcw.im/ufc"
)

var spaceReg = regexp.MustCompile(`\s+`)

func nowTimestamp() int64 {
	return time.Now().Unix()
}

func runCmd(name string, args ...string) (exitCode int, out string, err error) {
	cmd := exec.Command(name, args...)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	return cmd.ProcessState.ExitCode(), string(data), nil
}

type disk struct {
	total, used, available, percent int
}

// diskInfo returns the usage information of the disk where a directory is located
func diskInfo(dirpath string) (info disk, err error) {
	code, out, err := runCmd("df", "--output=size,used,avail,pcent", dirpath)
	if code != 0 || err != nil {
		return
	}
	di := spaceReg.Split(strings.TrimSpace(out), -1)
	total, err := strconv.Atoi(di[4])
	if err != nil {
		return
	}
	used, err := strconv.Atoi(di[5])
	if err != nil {
		return
	}
	avai, err := strconv.Atoi(di[6])
	if err != nil {
		return
	}
	percent, err := strconv.Atoi(strings.TrimSuffix(di[7], "%"))
	if err != nil {
		return
	}
	return disk{total, used, avai, percent}, nil
}

// diskRate returns the usage rate of the disk where the directory is located
func diskRate(dirpath string) (percent int, err error) {
	info, err := diskInfo(dirpath)
	if err != nil {
		return
	}
	return info.percent, nil
}

// memRate returns system memory usage
func memRate() (percent float64, err error) {
	f := "/proc/meminfo"
	if !ufc.IsFile(f) {
		err = errors.New("not found meminfo")
		return
	}
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return
	}
	mi := strings.Split(strings.TrimSpace(string(raw)), "\n")
	mem := make(map[string]int)
	for _, m := range mi {
		if m == "" {
			continue
		}
		mi := strings.Split(m, ":")
		k := strings.TrimSpace(mi[0])
		if k == "SwapCached" {
			break
		}
		v := strings.TrimSpace(mi[1])
		v = strings.Trim(v, "kB ")
		s, e := strconv.Atoi(v)
		if e != nil {
			err = e
			return
		}
		mem[k] = s
	}
	used := mem["MemTotal"] - mem["MemFree"] - mem["Buffers"] - mem["Cached"]
	percent = float64(used) / float64(mem["MemTotal"]) * 100
	p := fmt.Sprintf("%.2f", percent)
	return strconv.ParseFloat(p, 64)
}

// loadStat return load value in 5 minutes(float)
func loadStat() (loadavg5 float64, err error) {
	f := "/proc/loadavg"
	if !ufc.IsFile(f) {
		err = errors.New("not found loadavg")
		return
	}
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return
	}
	s := strings.Split(strings.TrimSpace(string(raw)), " ")[1]
	return strconv.ParseFloat(s, 64)
}

// getDirSize returns the total size of a directory
func getDirSize(dirpath string) (int64, error) {
	var size int64
	err := filepath.Walk(dirpath, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// makeTarFile compress all files in a directory
func makeTarFile(tarFilename, tarPath string, exclude []string) (err error) {
	// 创建文件
	fw, err := os.Create(tarFilename)
	if err != nil {
		return
	}
	defer fw.Close()

	// 创建 Tar.Writer 结构
	tw := tar.NewWriter(fw)
	defer tw.Close()

	// 递归处理目录及目录下的所有文件和目录
	return filepath.Walk(tarPath, func(fileName string, fi os.FileInfo, err error) error {
		// 因为这个闭包会返回个 error ，所以先要处理一下这个
		if err != nil {
			return err
		}

		// 如果文件名后缀（如 .gz .xxx）在排除列表中，则不压缩
		if ufc.StrInSlice(path.Ext(fileName), exclude) {
			return nil
		}

		// 判断下文件是否是标准文件，如果不是就不处理了，如：目录
		if !fi.Mode().IsRegular() {
			return nil
		}

		// 这里就不需要我们自己再 os.Stat 了，它已经做好了，我们直接使用 fi 即可
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		// 处理下 hdr 中的 Name，因为默认文件名不带路径，覆盖并去除首部/
		hdr.Name = strings.TrimPrefix(fileName, string(filepath.Separator))

		// 写入文件信息
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		// 打开文件
		fr, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer fr.Close()

		// copy 文件数据到 tw
		_, err = io.Copy(tw, fr)
		if err != nil {
			return err
		}

		return nil
	})
}

func SHA1(text string) string {
	if text == "" {
		return ""
	}
	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
