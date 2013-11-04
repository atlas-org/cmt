package cmt

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonuts/logger"
)

type Cmt struct {
	env *Setup // environment configured for cmt
	bin string // path to cmt.exe
	msg *logger.Logger
}

func New(env *Setup) (*Cmt, error) {
	var err error
	if env == nil {
		verbose := false
		env, err = newSetup("<local>", "", "", verbose)
		if err != nil {
			return nil, err
		}
	}

	out, err := env.sh.Run("which", "cmt.exe")
	if err != nil {
		return nil, err
	}
	bin := string(bytes.Trim(out, "\n"))

	cmt := &Cmt{
		env: env,
		bin: bin,
		msg: logger.New("cmt"),
	}

	dag, err := cmt.ProjectsDag()
	if err != nil {
		return nil, err
	}
	if len(dag) <= 0 {
		return nil, fmt.Errorf("cmt: no projects found. corrupted CMT environment ?")
	}
	return cmt, nil
}

// CheckOut checks out the package 'pkg' with revision 'version'.
//  pkg is the fullname of the package. e.g. Control/AthenaKernel
//  version can be empty to mean the HEAD or trunk or master
func (cmt *Cmt) CheckOut(pkg, version string) error {
	args := []string{"co", pkg}
	if version != "" {
		args = []string{"co", "-r", version, pkg}
	}
	out, err := cmt.env.sh.Run(cmt.bin, args...)
	if err != nil {
		cmt.errorf(
			"Problem running 'cmt co'. Failed to issue %s %s\n",
			cmt.bin,
			strings.Join(args, " "),
		)
		cmt.warnf("%v\n", string(out))
		return err
	} else {
		cmt.infof("## %s [OK]\n", pkg)
	}

	return err
}

func (cmt *Cmt) errorf(format string, args ...interface{}) {
	cmt.msg.Errorf(format, args...)
}

func (cmt *Cmt) warnf(format string, args ...interface{}) {
	cmt.msg.Warnf(format, args...)
}

func (cmt *Cmt) infof(format string, args ...interface{}) {
	cmt.msg.Infof(format, args...)
}

func (cmt *Cmt) debugf(format string, args ...interface{}) {
	cmt.msg.Debugf(format, args...)
}

// PackageVersion returns the package version in the current release
func (cmt *Cmt) PackageVersion(pkg string) string {
	args := []string{"show", "versions", pkg}
	cmt.debugf("running %v...\n", args)
	out, err := cmt.env.sh.Run(cmt.bin, args...)
	if err != nil {
		cmt.errorf(
			"Problem running PackageVersion. Failed to issue %s %s\n",
			cmt.bin,
			strings.Join(args, " "),
		)
		cmt.errorf("%v\n", string(out))
		return ""
	} else {
		cmt.debugf("## --- output ---:\n%v\n", string(out))
	}

	version := []byte("")
	area := cmt.env.sh.Getenv("TestArea")
	cmt.debugf("TestArea: %q\n", area)
	for _, line := range bytes.Split(out, []byte("\n")) {
		if area != "" && bytes.Index(line, []byte(area)) != -1 {
			continue
		}
		version = bytes.Split(line, []byte(" "))[1]
		break
	}
	return string(version)
}

// Show runs the 'cmt show xxx' command
func (cmt *Cmt) Show(args ...string) ([]byte, error) {
	cmt.debugf("running cmt show %v...\n", args)
	cmdargs := append([]string{"show"}, args...)
	out, err := cmt.env.sh.Run(cmt.bin, cmdargs...)
	if err != nil {
		cmt.errorf(
			"Problem running Show. Failed to issue %s %s\n",
			cmt.bin,
			strings.Join(cmdargs, " "),
		)
		cmt.errorf("%v\n", string(out))
		return nil, err
	} else {
		cmt.debugf("## --- output ---:\n%v\n", string(out))
	}

	return out, err
}

// Projects returns an unordered tree of all the projects
func (cmt *Cmt) Projects() (Projects, error) {

	out, err := cmt.Show("projects", "-xml")
	if err != nil {
		return nil, err
	}

	dec := xml.NewDecoder(bytes.NewBuffer(out))
	data := xmlTree{}
	err = dec.Decode(&data)
	if err != nil {
		cmt.errorf(
			"Problem decoding xml from 'cmt show projects -xml: %v\n",
			err,
		)
		return nil, err
	}

	projects := make(Projects)

	for _, proj := range data.Projects {
		pname := proj.Path
		p := NewProject(proj.Path, proj.Version)
		if p.Name == "CMTHOME" || p.Name == "CMTUSERCONTEXT" {
			// ignore that guy, it is a special (pain in the neck) one
			// see bug #75846
			// https://savannah.cern.ch/bugs/?75846
			continue
		}
		if proj.Current == "yes" {
			p.current = true
		}
		projects[pname] = &p
	}

	for _, xproj := range data.Projects {
		pname := xproj.Path
		proj := projects[pname]
		for _, client := range xproj.Clients {
			c := projects[client.Path]
			proj.Clients = append(proj.Clients, c)
		}
		for _, use := range xproj.Uses {
			u := projects[use.Path]
			proj.Uses = append(proj.Uses, u)
		}
	}
	return projects, nil
}

// ProjectsDag returns the directed-acyclic-graph of dependencies between projects
func (cmt *Cmt) ProjectsDag() (ProjectsDag, error) {
	projs, err := cmt.Projects()
	if err != nil {
		return nil, err
	}

	dag := make([]*Project, 0, len(projs))
	var root *Project
	nroots := 0
	for _, p := range projs {
		if p.current {
			nroots += 1
			root = p
		}
	}
	if nroots != 1 {
		return nil, fmt.Errorf(
			"cmt.dag: project tree inconsistency (found [%d] roots)",
			nroots,
		)
	}

	var visit func(p *Project, stack *[]*Project)
	visit = func(p *Project, stack *[]*Project) {
		if !has_project(*stack, p) {
			*stack = append(*stack, p)
		}
		for _, pp := range p.Clients {
			visit(pp, stack)
		}
	}
	visit(root, &dag)

	return ProjectsDag(dag), err
}

// Package returns a Cmt package by basename (or nil)
func (cmt *Cmt) Package(name string) (*Package, error) {
	dag, err := cmt.ProjectsDag()
	if err != nil {
		return nil, err
	}

	use := []byte("use ")
	for _, proj := range dag {
		fname := filepath.Join(
			proj.Path,
			project_release(proj),
			"cmt",
			"requirements",
		)
		if !path_exists(fname) {
			continue
		}
		f, err := os.Open(fname)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		scan := bufio.NewScanner(f)
		for scan.Scan() {
			bline := scan.Bytes()
			bline = bytes.Trim(bline, " \n")
			if !bytes.HasPrefix(bline, use) {
				continue
			}
			bline = bline[len(use):]
			if !bytes.HasPrefix(bline, []byte(name)) {
				continue
			}
			fields := make([]string, 0, 3)
			for _, tok := range bytes.Split(bline, []byte(" ")) {
				tok = bytes.Trim(tok, " \n")
				if len(tok) <= 0 {
					continue
				}
				fields = append(fields, string(tok))
			}
			switch len(fields) {
			case 2:
				return &Package{
					Name:    fields[0],
					Version: fields[1],
					Project: proj.Name,
				}, nil
			case 3:
				return &Package{
					Name:    filepath.Join(fields[2], fields[0]),
					Version: fields[1],
					Project: proj.Name,
				}, nil
			default:
				return nil, fmt.Errorf("cmt: malformed requirements file [%s]", fname)
			}
		}
		err = scan.Err()
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("cmt: package [%s] not found", name)
}

// EOF
