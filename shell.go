package cmt

import (
	"fmt"
	"io"
	"os/exec"
)

type Shell struct {
	cmd    *exec.Cmd
	Stdin  io.Writer
	Stdout io.Reader
}

func NewShell() (Shell, error) {
	cmd := exec.Command("/bin/sh")
	stdin, w := io.Pipe()
	cmd.Stdin = stdin

	r, stdout := io.Pipe()
	cmd.Stdout = stdout

	sh := Shell{
		cmd:    cmd,
		Stdin:  w,
		Stdout: r,
	}

	err := cmd.Start()
	if err != nil {
		return sh, err
	}
	return sh, nil
}

func (sh *Shell) Setenv(key, value string) error {
	_, err := sh.Stdin.Write([]byte(fmt.Sprintf("export %s=%s\n", key, value)))
	return err
}

func (sh *Shell) Getenv(key string) string {
	_, err := sh.Stdin.Write([]byte(fmt.Sprintf("echo ${%s}\n", key)))
	if err != nil {
		return ""
	}
	return "boo"
}

// EOF
