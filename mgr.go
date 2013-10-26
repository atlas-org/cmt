package cmt

// Mgr manages CMT environments
type Mgr struct {
	name string // project name
	topdir string // directory holding the whole project/workarea
	asetup_root string // path to asetup
	sh Shell // subshell where CMT is configured
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
