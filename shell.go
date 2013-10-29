package cmt

import (
	//"bytes"
	"fmt"
	"io"
	//"os"
	"os/exec"
	"strings"
)

type Shell struct {
	cmd    *exec.Cmd
	stdin  io.Writer
	stdout io.Reader
	resp   chan response
}

type response struct {
	buf []byte
	err error
}

func NewShell() (Shell, error) {
	cmd := exec.Command("/bin/sh")
	stdin, w := io.Pipe()
	cmd.Stdin = stdin

	r, stdout := io.Pipe()
	cmd.Stdout = stdout

	sh := Shell{
		cmd:    cmd,
		stdin:  w,
		stdout: r,
		resp:   make(chan response),
	}

	go func() {
		for {
			buf := make([]byte, 1024)
			//fmt.Fprintf(os.Stderr, "==>\n")
			n, err := sh.stdout.Read(buf)
			//fmt.Fprintf(os.Stderr, "==> %v (err=%v)\n", n, err)
			buf = buf[:n]
			//fmt.Fprintf(os.Stderr, "<<< [%v]\n", string(buf))
			sh.resp <- response{buf, err}
		}
	}()
	err := cmd.Start()
	if err != nil {
		return sh, err
	}
	return sh, nil
}

func (sh *Shell) Setenv(key, value string) error {
	_, err := sh.stdin.Write([]byte(fmt.Sprintf("export %s=%s\n", key, value)))
	return err
}

func (sh *Shell) Getenv(key string) string {
	_, err := sh.stdin.Write([]byte(fmt.Sprintf("echo ${%s}\n", key)))
	if err != nil {
		return ""
	}
	resp := <-sh.resp
	if resp.err != nil {
		return ""
	}
	out := string(resp.buf)
	out = strings.Trim(out, "\r\n")
	//fmt.Printf("::: %s [%v]\n", key, []byte(out))
	return out
}

func (sh *Shell) Source(script string) error {
	_, err := sh.stdin.Write([]byte(fmt.Sprintf(". %s; echo $?\n", script)))
	return err
}

// EOF
