package main

import (
	"testing"
)

func TestUtil(t *testing.T) {
	s1k := "hello world!"
	s1v := "430ce34d020724ed75a196dfc2ad67c77772d169"
	if SHA1(s1k) != s1v {
		t.Fatal("sha1 fail")
	}
}
