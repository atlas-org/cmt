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
	return &Cmt{
		env: env,
		bin: bin,
		msg: log.New(os.Stderr, "cmt:  ", 0),
	}, nil
}

// EOF
