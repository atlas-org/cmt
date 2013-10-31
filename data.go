package cmt

import (
	"path/filepath"
)

// Package represents a CMT package.
type Package struct {
	Name    string // package full name
	Version string // package version
	Project string // name of the project this package lives in
}

// Base returns the basename of this package
func (p *Package) Base() string {
	return filepath.Base(p.Name)
}

// Dir returns the dirname of this package
func (p *Package) Dir() string {
	return filepath.Dir(p.Name)
}

// Project represents a CMT project
type Project struct {
	Name    string // name of that project
	Version string // version of that project
	Path    string // path to where the project is installed
	Uses    []*Project
	Clients []*Project
}

func NewProject(path, version string) Project {
	return Project{
		Name:    filepath.Base(filepath.Dir(path)),
		Version: version,
		Path:    path,
		Uses:    make([]*Project, 0),
		Clients: make([]*Project, 0),
	}
}

func (p *Project) String() string {
	clients := make([]string, 0, len(p.Clients))
	for _, client := range p.Clients {
		clients = append(clients, client.Name)
	}
	uses := make([]string, 0, len(p.Uses))
	for _, use := range p.Uses {
		uses = append(uses, use.Name)
	}
	return fmt.Sprintf(
		"cmt.Project{name=%q, version=%q, clients=%v, uses=%v}",
		p.Name,
		p.Version,
		clients,
		uses,
	)
}

// Projects is the projects (dependency) tree
type Projects map[string]*Project

type xmlTree struct {
	XMLName xml.Name `xml:"projects"`

	Projects []*xmlProject `xml:"project"`
}

type xmlProject struct {
	Current string      `xml:"current,attr"`
	Name    string      `xml:"name"`
	Order   int         `xml:"order"`
	Version string      `xml:"version"`
	Path    string      `xml:"cmtpath"`
	Clients []xmlClient `xml:"clients>project"`
	Uses    []xmlUse    `xml:"uses>project"`
}

func (p *xmlProject) String() string {
	clients := make([]string, 0, len(p.Clients))
	for _, client := range p.Clients {
		clients = append(clients, client.Name)
	}
	uses := make([]string, 0, len(p.Uses))
	for _, use := range p.Uses {
		uses = append(uses, use.Name)
	}
	return fmt.Sprintf(
		"{name=%q, version=%q, path=%q, clients=%v, uses=%v}",
		p.Name,
		p.Version,
		p.Path,
		clients,
		uses,
	)
}

type xmlClient struct {
	Order   int    `xml:"order"`
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Path    string `xml:"cmtpath"`
}

type xmlUse struct {
	Order   int    `xml:"order"`
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Path    string `xml:"cmtpath"`
}

// EOF
