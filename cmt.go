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
	env Env    // environment configured for cmt
	bin string // path to cmt.exe
	msg *logger.Logger
}

func New(setup *Setup) (*Cmt, error) {
	var err error
	if setup == nil {
		verbose := false
		setup, err = newSetup("<local>", "", "", verbose)
		if err != nil {
			return nil, err
		}
		pwd := setup.topdir
		pwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
		err = setup.env.Chdir(pwd)
		if err != nil {
			return nil, err
		}
	}

	out, err := setup.env.Command("which", "cmt.exe").CombinedOutput()
	if err != nil {
		return nil, err
	}
	bin := string(bytes.Trim(out, "\n"))

	cmt := &Cmt{
		env: setup.env,
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
	out, err := cmt.env.Command(cmt.bin, args...).CombinedOutput()
	if err != nil {
		cmt.errorf(
			"Problem running 'cmt co'. Failed to issue %s %s\n",
			cmt.bin,
			strings.Join(args, " "),
		)
		cmt.warnf("%v\n", string(out))
		return err
	} else {
		cmt.debugf("%s [OK]\n", pkg)
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
	out, err := cmt.env.Command(cmt.bin, args...).CombinedOutput()
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
	area := cmt.env.Getenv("TestArea")
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
	out, err := cmt.env.Command(cmt.bin, cmdargs...).CombinedOutput()
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
		p.order = proj.Order
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
	omap := make(map[int]*Project, len(projs))
	var root *Project
	for _, p := range projs {
		omap[p.order] = p
	}
	for i, p := range omap {
		if len(omap[i].Clients) <= 0 && len(omap[i].Uses) > 0 {
			root = p
			break
		}
	}

	if root == nil {
		return nil, fmt.Errorf(
			"cmt.dag: project tree inconsistency (did not find any suitable root)",
		)
	} else {
		cmt.debugf("root=%s\n", root.Name)
	}

	var visit func(p *Project, stack *[]*Project)
	visit = func(p *Project, stack *[]*Project) {
		if !has_project(*stack, p) {
			*stack = append(*stack, p)
		}
		for _, pp := range p.Uses {
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
			bline = bytes.Trim(bline[len(use):], " \n")
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

// LatestPackageTag returns the most recent SVN tag of `pkg`
func (cmt *Cmt) LatestPackageTag(pkg string) (string, error) {
	svnroot := cmt.env.Getenv("SVNROOT")
	if svnroot == "" {
		return "", fmt.Errorf("cmt: SVNROOT not set")
	}
	args := []string{"ls", strings.Join([]string{svnroot, pkg, "tags"}, "/")}
	if strings.HasPrefix(pkg, "Gaudi") {
		svnroot = cmt.env.Getenv("GAUDISVN")
		if svnroot == "" {
			svnroot = "http://svnweb.cern.ch/guest/gaudi"
		}
		args = []string{"ls", strings.Join([]string{svnroot, "Gaudi", "tags", pkg}, "/")}
	}
	cmt.debugf("running svn %v...\n", args)
	bout, err := cmt.env.Command("svn", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cmt: error running svn %v:\nout:\n%v\nerr: %v",
			args,
			string(bout),
			err,
		)
	}
	tags := []string{}
	for _, bline := range bytes.Split(bout, []byte("\n")) {
		bline = bytes.Trim(bline, " \n")
		tags = append(tags, string(bline))
	}
	if len(tags) <= 0 {
		return "", fmt.Errorf("cmt: empty %s SVN directory", args[1])
	}

	bname := filepath.Base(pkg)

	// enforce atlas convention of tags (pkgname-xx-yy-zz-ww)
	tag := ""
	for _, t := range tags {
		if strings.HasPrefix(t, bname+"-") {
			tag = strings.Trim(t, " /\r\n")
		}
	}
	return tag, err
}

// EOF
