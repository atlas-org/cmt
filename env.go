package cmt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Env models a CMT environment
type Env struct {
	env map[string]string // environment for the subprocess
}

// Environ returns a copy of strings representing the environment, in the
// form "key=value".
func (env Env) Environ() []string {
	out := make([]string, 0, len(env.env))
	for k, v := range env.env {
		out = append(out, fmt.Sprintf("%s=%q", k, v))
	}
	return out
}

// EnvMap returns a copy of pairs key:value representing the environment.
func (env Env) EnvMap() map[string]string {
	out := make(map[string]string, len(env.env))
	for k, v := range env.env {
		out[k] = v
	}
	return out
}

func newEnv(env []string) Env {
	if len(env) == 0 || env == nil {
		env = os.Environ()
	}
	dict := make(map[string]string, len(env))
	for _, kv := range env {
		toks := strings.SplitN(kv, "=", 2)
		k := toks[0]
		v := toks[1]
		dict[k] = v
	}
	return Env{env: dict}
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not
// present.
func (env Env) Getenv(key string) string {
	v, _ := env.env[key]
	return v
}

// Getwd returns a rooted path name corresponding to the current directory.
// If the current directory can be reached via multiple paths (due to
// symbolic links), Getwd may return any one of them.
func (env Env) Getwd() (pwd string, err error) {
	pwd, _ = env.env["PWD"]
	return
}

// Setenv sets the value of the environment variable named by the key. It
// returns an error, if any.
func (env Env) Setenv(key, value string) error {
	env.env[key] = value
	return nil
}

// Chdir changes the current working directory to the named directory. If
// there is an error, it will be of type *PathError.
func (env Env) Chdir(dir string) error {
	env.env["PWD"] = dir
	return nil
}

// Command returns the Cmd struct to execute the named program with the
// given arguments.
func (env Env) Command(cmd string, args ...string) *exec.Cmd {
	if strings.Index(cmd, "/") == -1 {
		var err error
		cmd, err = env.LookPath(cmd)
		if err != nil {
			fmt.Printf("PATH=%q\n", env.env["PATH"])
			panic(err)
			return nil
		}
	}
	c := exec.Command(cmd, args...)
	c.Dir = env.env["PWD"]
	c.Env = env.Environ()
	return c
}

// LookPath searches for an executable binary named file
// in the directories named by the PATH environment variable.
// If file contains a slash, it is tried directly and the PATH is not consulted.
func (env Env) LookPath(file string) (string, error) {

	if strings.Contains(file, "/") {
		err := findExecutable(file)
		if err == nil {
			return file, nil
		}
		return "", &exec.Error{file, err}
	}
	pathenv := env.Getenv("PATH")
	if pathenv == "" {
		return "", &exec.Error{file, exec.ErrNotFound}
	}
	for _, dir := range strings.Split(pathenv, ":") {
		if dir == "" {

			dir = "."
		}
		path := dir + "/" + file
		if err := findExecutable(path); err == nil {
			return path, nil
		}
	}
	return "", &exec.Error{file, exec.ErrNotFound}
}

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}

// source sources the script given as argument and updates the environment
func (env *Env) asource(script string, args ...string) error {
	tmp, err := ioutil.TempDir("", "atl-cmt-env-source-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// FIXME: if script isn't absolute, should this be relative to
	// the parent process' dir or the env's pwd ?
	script, err = filepath.Abs(script)
	if err != nil {
		return err
	}

	fname := filepath.Join(tmp, "env.sh")
	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	markup := []byte("=== GO_SHELL ===")
	_, err = f.WriteString(fmt.Sprintf(`#!/bin/sh
# source the script
source %s %s

# markup
echo %q
/usr/bin/printenv
echo %q

## EOF
`,
		script,
		strings.Join(args, " "),
		string(markup),
		string(markup),
	))
	if err != nil {
		return err
	}

	cmd := env.Command("/bin/sh", fname)
	bout, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmt.env: %v\n%v", err, string(bout))
	}

	start_idx := bytes.Index(bout, markup)
	if start_idx == -1 {
		return fmt.Errorf("cmt.env: failed to extract source'd environment")
	}
	penv := bout[start_idx+len(markup):]

	stop_idx := bytes.Index(penv, markup)
	if stop_idx == -1 {
		return fmt.Errorf("cmt.env: failed to extract source'd environment")
	}
	penv = penv[:stop_idx]

	blines := bytes.Split(bytes.Trim(penv, "\n"), []byte("\n"))
	for _, kv := range blines {
		toks := strings.SplitN(string(kv), "=", 2)
		key := toks[0]
		val := toks[1]
		if key == "_" {
			continue
		}
		env.env[key] = val
	}

	return nil
}

// EOF
