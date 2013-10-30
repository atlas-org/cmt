package cmt

import (
	"bytes"

	"github.com/atlas-org/shell"
)

type Cmt struct {
	sh  shell.Shell // environment configured for cmt
	bin string      // path to cmt.exe
}

func New(sh shell.Shell) (*Cmt, error) {
	out, err := sh.Run("which", "cmt.exe")
	if err != nil {
		return nil, err
	}
	bin := string(bytes.Trim(out, "\n"))
	return &Cmt{
		sh:  sh,
		bin: bin,
	}, nil
}

// EOF
