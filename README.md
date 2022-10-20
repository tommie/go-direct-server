# Go Direct Server

Command `godirectserverd` runs an HTTP server responding to ?go-get=1 queries to resolve VCS URLs from Go module paths.
See https://go.dev/ref/mod#vcs-find.

It is meant to be low-configuration.
Because it is expected to run behind a reverse proxy, it does not support TLS.

## Running

    go run github.com/tommie/go-direct-server/cmd/godirectserverd@latest -listen-addr localhost:8080

However, you probably want to put that in a container.

This will start a server that uses GitHub-style resolution, so module
path `localhost/a/b/c` will be resolved to `git ssh://git@localhost/a/b`.

## Rules

The `-rule-file` flag takes a path to a file with lines containing

    modpath-pattern  root-tmpl vcs-tmpl repourl-tmpl

* The `modpath-pattern` is a Go module path regexp like `golang.org/(.*)`.
* The `root-tmpl` is the repository's module root, e.g. `golang.org/$1`.
* The `vcs-tmpl` is the VCS to use, e.g. `git`.
* The `repourl-tmpl` is the repository URL, e.g. `ssh://git@github.com/$1`.

Empty lines and those starting with hash are ignored.

## Configuration

These flags are available:

* **-host-header** sets the HTTP request header to read the server hostname from.
  This is useful behind a reverse proxy that sets e.g. `X-Forwarded-Host`.
* **-listen-addr** sets the TCP address to listen for HTTP connections on.
* **-rule-file** sets the path of the rule file.
