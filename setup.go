package cmt

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/atlas-org/shell"
)

// Setup manages a CMT environment
type Setup struct {
	name    string      // project name
	topdir  string      // directory holding the whole project/workarea
	asetup  string      // path to asetup.sh
	sh      shell.Shell // subshell where CMT is configured
	verbose bool
}

// NewSetup returns a Cmt Setup configured with the given tags
func NewSetup(tags string, verbose bool) (*Setup, error) {
	project := os.Getenv("AtlasProject")
	if project == "" {
		project = "AtlasOffline"
	}

	asetup_root := "/afs/cern.ch/atlas/software/dist/AtlasSetup"

	return newSetup(project, asetup_root, tags, verbose)
}

func newSetup(project, asetup_root, tags string, verbose bool) (*Setup, error) {

	topdir, err := ioutil.TempDir("", "atl-cmt-mgr-")
	if err != nil {
		return nil, err
	}

	sh, err := shell.New()
	if err != nil {
		return nil, err
	}

	mgr := &Setup{
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

func (mgr *Setup) init(tags string) error {
	var err error
	err = mgr.create_asetup_cfg(tags)
	if err != nil {
		return err
	}
	return err
}

func (mgr *Setup) create_asetup_cfg(tags string) error {
	var err error
	err = mgr.sh.Chdir(mgr.topdir)
	if err != nil {
		return err
	}
	fname := filepath.Join(mgr.topdir, ".asetup.cfg")
	if mgr.verbose {
		fmt.Printf("cmt: create [%s]...\n", fname)
	}
	cfg, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer cfg.Close()
	_, err = cfg.WriteString(`
[defaults]
opt = True
lang = C
hastest = True  ## to prepend pwd to cmtpath
pedantic = True
runtime = True
setup = True
os = slc6
save = True
testarea=<pwd>
`)
	if err != nil {
		return err
	}
	err = cfg.Sync()
	if err != nil {
		return err
	}
	err = cfg.Close()
	if err != nil {
		return err
	}
	// source it
	args := []string{"--input=" + fname, tags}
	if mgr.verbose {
		fmt.Printf("cmt: sourcing 'asetup'...\n")
	}
	err = mgr.sh.Source(mgr.asetup, args...)
	if err != nil {
		return fmt.Errorf("cmt: error sourcing 'asetup': %v", err)
	}

	if mgr.verbose {
		fmt.Printf("cmt: running 'cmt show path'...\n")
	}
	out, err := mgr.sh.Run("cmt", "show", "path")
	if err != nil {
		return fmt.Errorf("cmt: error running 'cmt show path': %v", err)
	}
	if mgr.verbose {
		fmt.Printf("cmt: 'cmt show path':\n%v\n===EOF===\n", string(out))
	}

	return err
}

func (mgr *Setup) Delete() error {
	return combineErrors(
		os.RemoveAll(mgr.topdir),
		mgr.sh.Delete(),
	)
}

// func (mgr *Setup) Shell() shell.Shell {
// 	return mgr.sh
// }

type merror struct {
	errs []error
}

func (err merror) Error() string {
	o := make([]string, 0, len(err.errs))
	for i, e := range err.errs {
		o = append(
			o,
			fmt.Sprintf("[%d]: %v", i, e),
		)
	}
	return strings.Join(o, "\n")
}

func combineErrors(errs ...error) error {
	stack := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			stack = append(stack, err)
		}
	}
	if len(stack) == 0 {
		return nil
	}
	return merror{stack}
}

// EOF
