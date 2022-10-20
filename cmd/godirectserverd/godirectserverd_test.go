package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"testing"
)

func TestServe(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer l.Close()

	go serve(l, defaultRules, "host")

	t.Run("repoRootFound", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/a/b", l.Addr()))
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		defer resp.Body.Close()

		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("Get StatusCode: got %v, want %v", resp.StatusCode, want)
		}

		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		got := contentRE.FindStringSubmatch(string(bs))
		if want := "127.0.0.1/a/b git ssh://git@127.0.0.1/a/b"; len(got) > 1 && got[1] != want {
			t.Errorf("Get content: got %q, want %q", got[1], want)
		}
	})

	t.Run("repoSubmoduleFound", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/a/b/c", l.Addr()))
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		defer resp.Body.Close()

		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("Get StatusCode: got %v, want %v", resp.StatusCode, want)
		}

		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		got := contentRE.FindStringSubmatch(string(bs))
		if want := "127.0.0.1/a/b git ssh://git@127.0.0.1/a/b"; len(got) > 1 && got[1] != want {
			t.Errorf("Get content: got %q, want %q", got[1], want)
		}
	})

	t.Run("rootNotFound", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://%s/", l.Addr()))
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		defer resp.Body.Close()

		if want := http.StatusNotFound; resp.StatusCode != want {
			t.Errorf("Get StatusCode: got %v, want %v", resp.StatusCode, want)
		}
	})

	t.Run("nonStandardHost", func(t *testing.T) {
		l, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			t.Fatalf("Listen failed: %v", err)
		}
		defer l.Close()

		go serve(l, defaultRules, "x-forwarded-host")

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/a/b", l.Addr()), nil)
		if err != nil {
			t.Fatalf("NewRequest failed: %v", err)
		}
		req.Header.Set("X-Forwarded-Host", "ahostname:1234")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		defer resp.Body.Close()

		if want := http.StatusOK; resp.StatusCode != want {
			t.Errorf("Get StatusCode: got %v, want %v", resp.StatusCode, want)
		}

		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		got := contentRE.FindStringSubmatch(string(bs))
		if want := "ahostname/a/b git ssh://git@ahostname/a/b"; len(got) > 1 && got[1] != want {
			t.Errorf("Get content: got %q, want %q", got[1], want)
		}
	})
}

var contentRE = regexp.MustCompile(`content="([^"]+)"`)
