// web tools

package main

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type download struct {
	Uifn           string  `json:"uifn"`
	UifnKey        string  `json:"uifnKey"`
	Site           uint8   `json:"site"`
	BoardId        string  `json:"board_id"`
	Ctime          uint    `json:"ctime"`
	Etime          uint    `json:"etime"`
	BoardPins      string  `json:"board_pins"`
	downloads      []pin   // Go type after board_pins parsing
	MAXBoardNumber uint    `json:"MAX_BOARD_NUMBER"`
	CallbackURL    string  `json:"CALLBACK_URL"`
	DiskLimit      float64 `json:"DISKLIMIT"`
}

type clean struct {
	Uifn        string `json:"uifn"`
	CallbackURL string `json:"CALLBACK_URL"`
}

type pin struct {
	Name string `json:"imgName"`
	URL  string `json:"imgUrl"`
}

type eres struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func splitPins(arr []pin, num int) (segmens [][]pin, err error) {
	// num is the number of splits
	max := len(arr)
	if max < num {
		err = errors.New("out of slice size")
		return
	}
	quantity := max / num
	end := 0
	for i := 1; i <= num; i++ {
		qu := i * quantity
		if i != num {
			segmens = append(segmens, arr[i-1+end:qu])
		} else {
			segmens = append(segmens, arr[i-1+end:])
		}
		end = qu - i
	}
	return
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusBadRequest
	msg := err.Error()
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message.(string)
	}
	c.JSON(code, eres{-1, msg})
}

func signatureRequired(c echo.Context) error {
	signature := c.QueryParam("signature")
	timestamp := c.QueryParam("timestamp")
	nonce := c.QueryParam("nonce")
	if signature == "" || timestamp == "" || nonce == "" {
		return errors.New("invalid param")
	}
	err := checkTimestamp(timestamp)
	if err != nil {
		return err
	}
	if passed := checkSignature(signature, timestamp, nonce); passed {
		return nil
	}
	return errors.New("signature verification failed")
}

func checkTimestamp(reqTimestamp string) error {
	if len(reqTimestamp) != 10 {
		return errors.New("invalid timestamp")
	}
	timestamp, err := strconv.Atoi(reqTimestamp)
	if err != nil {
		return err
	}
	rt := int64(timestamp)
	nt := nowTimestamp()
	if (rt <= nt || rt-10 <= nt) && (rt+300 >= nt) {
		return nil
	}
	return errors.New("check timestamp fail")
}

func checkSignature(signature, timestamp, nonce string) bool {
	args := []string{token, timestamp, nonce}
	sort.Strings(args)
	mysig := SHA1(strings.Join(args, ""))
	return mysig == signature
}

func httpGet(url string, headers map[string]string, timeout time.Duration) (resp *http.Response, err error) {
	var client = &http.Client{Timeout: timeout}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return client.Do(req)
}

func httpPost(url string, data map[string]string) (resp *http.Response, err error) {
	var post http.Request
	post.ParseForm()
	for k, v := range data {
		post.Form.Add(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(
		"POST", url, strings.NewReader(post.Form.Encode()),
	)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "tdi/v"+version)

	return client.Do(req)
}
