package main

import (
	"context"
	"errors"
	"regexp"
)

// Resolver resolves VCS repositories from a Go module path using a
// simple RE2 rule set.
type Resolver struct {
	Rules []*Rule
}

// A Rule encodes a match for a repository type. Templates may contain
// `$1` parameters, expanded from capture groups in the regular
// expression.
type Rule struct {
	ModPath *regexp.Regexp

	RootTmpl    string
	VCSTmpl     string
	RepoURLTmpl string
}

func (res *Resolver) Resolve(ctx context.Context, modPath string) (*Record, error) {
	for _, r := range res.Rules {
		m := r.ModPath.FindStringSubmatchIndex(modPath)
		if m != nil {
			return &Record{
				Root:    string(r.ModPath.ExpandString(nil, r.RootTmpl, modPath, m)),
				VCS:     VCS(r.ModPath.ExpandString(nil, r.VCSTmpl, modPath, m)),
				RepoURL: string(r.ModPath.ExpandString(nil, r.RepoURLTmpl, modPath, m)),
			}, nil
		}
	}

	return nil, ErrNotFound
}

var ErrNotFound = errors.New("module not found")

// A Record is the result of resolving a module path.
type Record struct {
	Root    string
	VCS     VCS
	RepoURL string
}

// VCS describes the type of version control system.
type VCS string

const (
	Bazaar     VCS = "bzr"
	Fossil     VCS = "fossil"
	Git        VCS = "git"
	Mercurial  VCS = "hg"
	Subversion VCS = "svn"
)
