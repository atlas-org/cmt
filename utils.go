package cmt

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonuts/logger"
)

func path_exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// project_release returns the name of the package holding
// the list of packages defining the release
func project_release(p *Project) string {
	n, ok := map[string]string{
		"LCGCMT":      "LCG_Release",
		"dqm-common":  "DQMCRelease",
		"tdaq-common": "TDAQCRelease",
	}[p.Name]
	if !ok {
		n = p.Name + "Release"
	}
	return n
}

// has_project returns true if p is in the slice
func has_project(projs []*Project, p *Project) bool {
	for _, pp := range projs {
		if pp == p {
			return true
		}
	}
	return false
}

// extract_uses returns the list of packages a given requirements file uses
func extract_uses(fname string, msg *logger.Logger) ([]Package, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pkgs := make([]Package, 0, 2)
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		line = strings.Trim(line, " \r\n\t")
		if !strings.HasPrefix(line, "use ") {
			continue
		}
		pkg := strings.Trim(line[len("use "):], " \r\n\t")
		toks := make([]string, 0)
		for _, tok := range strings.Split(pkg, " ") {
			tok = strings.Trim(tok, " \t")
			if tok != "" {
				toks = append(toks, tok)
			}
		}
		if len(toks) >= 1 {
			pkg_name := strings.Trim(toks[0], " ")
			pkg_vers := "*"
			if len(toks) >= 2 {
				pkg_vers = strings.Trim(toks[1], " ")
			}
			pkg_path := ""
			if len(toks) >= 3 {
				pkg_path = strings.Trim(toks[2], " ")
			}
			msg.Debugf("found [%s] [%s] [%s]\n", pkg_name, pkg_vers, pkg_path)
			pkg := Package{
				Name:    filepath.Join(pkg_path, pkg_name),
				Version: pkg_vers,
				Project: "",
			}
			pkgs = append(pkgs, pkg)
		} else {
			msg.Errorf("unexpected line content: %q\n", line)
			return nil, fmt.Errorf("invalid requirement file [%s]", fname)
		}
	}
	return pkgs, err
}

// EOF
