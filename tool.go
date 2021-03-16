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
	"syscall"
	"time"

	"tcw.im/ufc"
)

var spaceReg = regexp.MustCompile(`\s+`)

type sysinfo struct {
	Load5     float64 // 5,minute load averages
	TotalRam  uint64  // total usable main memory size [kB]
	FreeRam   uint64  // available memory size [kB]
	SharedRam uint64  // amount of shared memory [kB]
	BufferRam uint64  // memory used by buffers [kB]
}

func cwd() string {
	pwd, _ := os.Getwd()
	return pwd
}

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

// diskRate returns the usage rate of the disk where the directory is located
func diskRate(volumePath string) (percent float64, err error) {
	var fs syscall.Statfs_t
	err = syscall.Statfs(volumePath, &fs)
	if err != nil {
		return
	}
	Size := fs.Blocks * uint64(fs.Bsize)
	Free := fs.Bfree * uint64(fs.Bsize)
	Used := Size - Free

	pct := float64(Used) / float64(Size) * 100
	return strconv.ParseFloat(fmt.Sprintf("%.2f", pct), 64)
}

func newSysinfo() (sis *sysinfo, err error) {
	si := new(syscall.Sysinfo_t)
	err = syscall.Sysinfo(si)
	if err != nil {
		return
	}
	fmt.Printf("%+v\n", si)

	unit := uint64(si.Unit) * 1024 // kB
	scale := 65536.0               // magic

	sis.Load5 = float64(si.Loads[1]) / scale

	sis.TotalRam = uint64(si.Totalram) / unit
	sis.FreeRam = uint64(si.Freeram) / unit
	sis.SharedRam = uint64(si.Sharedram) / unit
	sis.BufferRam = uint64(si.Bufferram) / unit

	return
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
	fmt.Printf("%+v\n", mem)
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

// makeTarFile compress all files in a directory.
// Automatically delete after compression.
func makeTarFile(tarFilename, tarPath string, exclude []string) (err error) {
	if !strings.HasSuffix(tarFilename, ".tar") || !ufc.IsDir(tarPath) {
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
		if ufc.StrInSlice(path.Ext(fileName), exclude) {
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

// formatSize: format byte size as kilobytes, megabytes, gigabytes
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
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func SHA1(text string) string {
	if text == "" {
		return ""
	}
	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
