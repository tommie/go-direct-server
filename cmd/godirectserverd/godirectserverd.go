// Command godirectserverd runs an HTTP server responding to ?go-get=1
// queries to resolve VCS URLs from Go module paths.
//
// See https://go.dev/ref/mod#vcs-find.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	hostHeader = flag.String("host-header", "host", "The header to read the request hostname from.")
	listenAddr = flag.String("listen-addr", "localhost:0", "TCP address to listen for HTTP requests at.")
	ruleFile   = flag.String("rule-file", "", "File containing module path patterns and string templates.")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run() error {
	var rs []*Rule
	if *ruleFile == "" {
		log.Println("No -rule-file, using GitHub-style.")
		rs = defaultRules
	} else {
		var err error
		rs, err = loadRules(*ruleFile)
		if err != nil {
			return err
		}
	}

	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		return err
	}
	defer l.Close()

	return serve(l, rs, *hostHeader)
}

// serve runs the HTTP server on the given listener with the given
// rule set.
func serve(l net.Listener, rs []*Rule, hostHeader string) error {
	s := http.Server{
		Addr: l.Addr().String(),
		Handler: &goVCSHandler{
			r:          &Resolver{rs},
			hostHeader: strings.ToLower(hostHeader),
		},
	}

	log.Printf("Serving on %s with %d rule(s)...", s.Addr, len(rs))

	return s.Serve(l)
}

// loadRules reads rules from a file.
func loadRules(path string) ([]*Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readRules(f)
}

// readRules reads lines of rules from a reader.
func readRules(r io.Reader) ([]*Rule, error) {
	var rs []*Rule
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		ss := strings.Fields(line)
		if len(ss) != 4 {
			return nil, fmt.Errorf("expected four fields but got %d: %s", len(ss), line)
		}

		mpRE, err := regexp.Compile(ss[0])
		if err != nil {
			return nil, err
		}

		rs = append(rs, &Rule{
			ModPath:     mpRE,
			RootTmpl:    ss[1],
			VCSTmpl:     ss[2],
			RepoURLTmpl: ss[3],
		})
	}

	return rs, sc.Err()
}

var defaultRules = []*Rule{
	{
		ModPath:     regexp.MustCompile(`^([^/:]+)(?::[^/]+)?/([^/]+/[^/]+)(?:/.+)?`),
		RootTmpl:    "$1/$2",
		VCSTmpl:     "git",
		RepoURLTmpl: "ssh://git@$1/$2",
	},
}
