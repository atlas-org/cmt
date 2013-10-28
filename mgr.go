package cmt

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Mgr manages CMT environments
type Mgr struct {
	name        string // project name
	topdir      string // directory holding the whole project/workarea
	asetup_root string // path to asetup
	sh          Shell  // subshell where CMT is configured
}

// NewMgr returns a Cmt manager configured with the given tags
func NewMgr(tags string, verbose bool) (*Mgr, error) {
	project := os.Getenv("AtlasProject")
	if project == "" {
		project = "AtlasOffline"
	}

	cmt_root := os.Getenv("CMTROOT")
	if cmt_root == "" {
		return nil, fmt.Errorf("cmt: no CMTROOT env.var\n")
	}
	cmt_version := "v1r25"
	asetup_root := "/afs/cern.ch/atlas/software/dist/AtlasSetup"

	return newMgr(project, cmt_root, cmt_version, asetup_root, tags, verbose)
}

func newMgr(project, cmt_root, cmt_version, asetup_root, tags string, verbose bool) (*Mgr, error) {

	topdir, err := ioutil.TempDir("", "atl-cmt-mgr-")
	if err != nil {
		return nil, err
	}

	sh, err := NewShell()
	if err != nil {
		return nil, err
	}

	mgr := &Mgr{
		name:        project,
		topdir:      topdir,
		asetup_root: asetup_root,
		sh:          sh,
	}
	err = mgr.init()
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

// EOF
