// web api view

package main

import (
	"encoding/json"
	"errors"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"tcw.im/ufc"
)

func pingView(c echo.Context) error {
	if err := signatureRequired(c); err != nil {
		return err
	}
	load5, err := loadStat()
	if err != nil {
		return err
	}
	memp, err := memRate()
	if err != nil {
		return err
	}
	diskp, err := diskRate(dir)
	if err != nil {
		return err
	}
	info := make(map[string]interface{})
	info["code"] = 0
	info["version"] = version
	info["status"] = status
	info["memRate"] = memp
	info["diskRate"] = diskp
	info["loadFive"] = load5
	info["timestamp"] = time.Now().Unix()
	info["lang"] = runtime.Version()
	info["rqcount"] = -1
	info["rqfailed"] = -1
	info["goroutine"] = runtime.NumGoroutine()
	return c.JSON(200, info)
}

func downloadView(c echo.Context) error {
	if err := signatureRequired(c); err != nil {
		return err
	}
	data := &download{}
	if err := c.Bind(&data); err != nil {
		return err
	}

	if data.UifnKey == "" || data.Uifn == "" || data.BoardId == "" ||
		data.BoardPins == "" || data.Ctime == 0 || data.Etime == 0 {
		return errors.New("invalid param")
	}

	pins := make([]pin, 0)
	json.Unmarshal([]byte(data.BoardPins), &pins)
	if len(pins) < 1 {
		return errors.New("empty download")
	}
	data.downloads = pins

	// write to temp file
	simple := clean{data.Uifn, data.CallbackURL}
	if err := serialize(simple, data.Uifn); err != nil {
		return err
	}

	go downloadBoard(data)

	return c.JSONBlob(201, []byte(`{"code":0,"msg":"downloading"}`))
}

func sendfileView(c echo.Context) error {
	name := c.Param("filename")
	if name == "" || !strings.HasPrefix(name, "hb_") {
		return c.String(400, "illegal filename")
	}
	f := filepath.Join(dir, path.Clean(name))
	if !ufc.IsFile(f) {
		return c.String(404, "not found")
	}
	return c.Attachment(f, name)
}
