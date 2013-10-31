package cmt

import (
	"bytes"
	"encoding/xml"
	"log"
	"os"
	"strings"
)

type Cmt struct {
	env *Setup // environment configured for cmt
	bin string // path to cmt.exe
	msg *log.Logger
}

func New(env *Setup) (*Cmt, error) {
	out, err := env.sh.Run("which", "cmt.exe")
	if err != nil {
		return nil, err
	}
	bin := string(bytes.Trim(out, "\n"))

	cmt := &Cmt{
		env: env,
		bin: bin,
		msg: log.New(os.Stderr, "cmt:  ", 0),
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
	cmt.msg.Printf("ERROR    "+format, args...)
}

func (cmt *Cmt) warnf(format string, args ...interface{}) {
	cmt.msg.Printf("WARNING  "+format, args...)
}

func (cmt *Cmt) infof(format string, args ...interface{}) {
	cmt.msg.Printf("INFO     "+format, args...)
}

func (cmt *Cmt) debugf(format string, args ...interface{}) {
	cmt.msg.Printf("DEBUG    "+format, args...)
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

// EOF
