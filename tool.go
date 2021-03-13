package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"tcw.im/ufc"
)

var spaceReg = regexp.MustCompile(`\s+`)

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
func diskInfo(path string) (info disk, err error) {
	code, out, err := runCmd("df", "--output=size,used,avail,pcent", path)
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
func diskRate(path string) (percent int, err error) {
	info, err := diskInfo(path)
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
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
