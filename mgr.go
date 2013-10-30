package cmt

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/atlas-org/shell"
)

// Mgr manages CMT environments
type Mgr struct {
	name    string      // project name
	topdir  string      // directory holding the whole project/workarea
	asetup  string      // path to asetup.sh
	sh      shell.Shell // subshell where CMT is configured
	verbose bool
}

// NewMgr returns a Cmt manager configured with the given tags
func NewMgr(tags string, verbose bool) (*Mgr, error) {
	project := os.Getenv("AtlasProject")
	if project == "" {
		project = "AtlasOffline"
	}

	asetup_root := "/afs/cern.ch/atlas/software/dist/AtlasSetup"

	return newMgr(project, asetup_root, tags, verbose)
}

func newMgr(project, asetup_root, tags string, verbose bool) (*Mgr, error) {

	topdir, err := ioutil.TempDir("", "atl-cmt-mgr-")
	if err != nil {
		return nil, err
	}

	sh, err := shell.New()
	if err != nil {
		return nil, err
	}

	mgr := &Mgr{
		name:    project,
		topdir:  topdir,
		asetup:  filepath.Join(asetup_root, "scripts", "asetup.sh"),
		sh:      sh,
		verbose: verbose,
	}
	err = mgr.init(tags)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func (mgr *Mgr) init() error {
	var err error
	return err
}

func (mgr *Mgr) create_asetup_cfg() error {
	var err error
	return err
}

func (mgr *Mgr) Delete() error {
	return mgr.sh.Delete()
}

// EOF
