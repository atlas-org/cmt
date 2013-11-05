package cmt

import (
	"encoding/json"
	"fmt"
	"io"
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
	remove  bool        // switch whether to remove or not the topdir
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

// NewSetupFromCache returns a Cmt setup from a previously cached environment
func NewSetupFromCache(fname, topdir string, verbose bool) (*Setup, error) {
	var err error
	remove := false
	if topdir == "" {
		topdir, err = ioutil.TempDir("", "atl-cmt-mgr-")
		if err != nil {
			return nil, err
		}
		remove = true
	}

	sh, err := shell.New()
	if err != nil {
		return nil, err
	}

	asetup_root := "/afs/cern.ch/atlas/software/dist/AtlasSetup"
	project := "AtlasInvalid"

	s := &Setup{
		name:    project,
		topdir:  topdir,
		remove:  remove,
		asetup:  filepath.Join(asetup_root, "scripts", "asetup.sh"),
		sh:      sh,
		verbose: verbose,
	}

	f, err := os.Open(fname)
	if err != nil {
		s.Delete()
		return nil, err
	}
	defer f.Close()

	err = s.Load(f)
	if err != nil {
		s.Delete()
		return nil, err
	}

	s.name = s.sh.Getenv("AtlasProject")
	s.asetup = s.sh.Getenv("AtlasSetup")

	err = s.init()
	if err != nil {
		s.Delete()
		return nil, err
	}

	return s, nil
}

func newSetup(project, asetup_root, tags string, verbose bool) (*Setup, error) {

	topdir, err := ioutil.TempDir("", "atl-cmt-mgr-")
	if err != nil {
		return nil, err
	}

	sh, err := shell.New()
	if err != nil {
		sh.Delete()
		return nil, err
	}

	s := &Setup{
		name:    project,
		topdir:  topdir,
		remove:  true,
		asetup:  filepath.Join(asetup_root, "scripts", "asetup.sh"),
		sh:      sh,
		verbose: verbose,
	}
	err = s.init()
	if err != nil {
		s.Delete()
		return nil, err
	}

	if asetup_root != "" {
		err = s.create_asetup_cfg(tags)
		if err != nil {
			s.Delete()
			return nil, err
		}
	}

	return s, nil
}

func (s *Setup) init() error {
	var err error

	err = s.sh.Chdir(s.topdir)
	if err != nil {
		return err
	}

	return err
}

func (s *Setup) create_asetup_cfg(tags string) error {
	var err error
	fname := filepath.Join(s.topdir, ".asetup.cfg")
	if s.verbose {
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
	if s.verbose {
		fmt.Printf("cmt: sourcing 'asetup %v'...\n", args)
	}
	err = s.sh.Source(s.asetup, args...)
	if err != nil {
		return fmt.Errorf("cmt: error sourcing 'asetup': %v", err)
	}

	if s.verbose {
		fmt.Printf("cmt: running 'cmt show path'...\n")
	}
	out, err := s.sh.Run("cmt", "show", "path")
	if err != nil {
		return fmt.Errorf("cmt: error running 'cmt show path': %v", err)
	}
	if s.verbose {
		fmt.Printf("cmt: 'cmt show path':\n%v\n===EOF===\n", string(out))
	}

	return err
}

func (s *Setup) Delete() error {
	var err error
	if s.remove {
		err = os.RemoveAll(s.topdir)
	}
	return combineErrors(
		err,
		s.sh.Delete(),
	)
}

// func (s *Setup) Shell() shell.Shell {
// 	return s.sh
// }

// Save encodes the current setup in `w` as a JSON dict
func (s *Setup) Save(w io.Writer) error {

	dict := make(map[string]string)
	for k, v := range s.EnvMap() {
		if k == "_" {
			continue
		}
		v = strings.Replace(v, s.topdir, "@@GO_CMT_TOPDIR@@", -1)
		dict[k] = v
	}

	data, err := json.MarshalIndent(&dict, "", "  ")
	_, err = w.Write(data)
	return err
}

// Load restores a setup from `r`
func (s *Setup) Load(r io.Reader) error {
	// save current workdir
	wd, err := s.sh.Getwd()
	if err != nil {
		return err
	}
	// restore workdir
	defer s.sh.Chdir(wd)

	tmp, err := ioutil.TempDir("", "atl-cmt-mgr-load-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	fname := filepath.Join(tmp, "store.cmt")
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(r)
	if dec == nil {
		return fmt.Errorf("cmt.setup: could not create JSON decoder")
	}

	data := make(map[string]string)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}

	for k, v := range data {
		v = strings.Replace(v, "@@GO_CMT_TOPDIR@@", s.topdir, -1)
		_, err = f.WriteString(fmt.Sprintf("export %s=%q\n", k, v))
		if err != nil {
			return err
		}
	}
	err = f.Close()
	if err != nil {
		return err
	}

	return s.sh.Source(fname)
}

func (s *Setup) EnvMap() map[string]string {
	dict := make(map[string]string)
	for _, env := range s.sh.Environ() {
		toks := strings.SplitN(env, "=", 2)
		k := toks[0]
		v := toks[1]
		dict[k] = v
	}
	return dict
}

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
