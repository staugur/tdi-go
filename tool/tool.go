package tool

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

// DiskInfo returns the usage information of the disk where a directory is located
func DiskInfo(path string) (info disk, err error) {
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

// DiskRate returns the usage rate of the disk where the directory is located
func DiskRate(path string) (percent int, err error) {
	info, err := DiskInfo(path)
	if err != nil {
		return
	}
	return info.percent, nil
}

// MemRate returns system memory usage
func MemRate() (percent float64, err error) {
	raw, err := ioutil.ReadFile("/proc/meminfo")
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

// LoadStat return load value in 5 minutes(float)
func LoadStat() (loadavg5 float64, err error) {
	raw, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}
	s := strings.Split(strings.TrimSpace(string(raw)), " ")[1]
	return strconv.ParseFloat(s, 64)
}

// GetDirSize returns the total size of a directory
func GetDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
