// web tools

package main

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func errView(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, `{"code":-1,"msg":"%s"}`, err.Error())
}

func errView500(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"code":500,"msg":"%s"}`, err.Error())
}

func errView400(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{"code":400,"msg":"bad request"}`))
}

func errView404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"code":404,"msg":"not found"}`))
}

func errView405(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"code":405,"msg":"method not allowed"}`))
}

func signatureRequired(r *http.Request) error {
	arg := r.URL.Query()
	signature := arg.Get("signature")
	timestamp := arg.Get("timestamp")
	nonce := arg.Get("nonce")
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
