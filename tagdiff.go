package cmt

import (
	"fmt"
)

// TagDiff returns the list of tag differences between 2 releases/nightlies
func TagDiff(old, new string, verbose bool) ([]string, error) {
	var err error
	diff := make([]string, 0)

	cmts := map[string]*Cmt{
		"old": nil,
		"new": nil,
	}
	tags := map[string]string{
		"old": old,
		"new": new,
	}

	type response struct {
		name string
		cmt *Cmt
		err error
	}

	ch := make(chan response)

	for name, tag := range tags {
		go func(tag, name string, ch chan response) {
			if verbose {
				fmt.Printf("::: setup %s env. [%s]...\n", name, tag)
			}
			env, err := NewSetup(tag, verbose)
			if err != nil {
				ch <- response{name, nil, err}
				return
			}
			cmt, err := New(env)
			if err != nil {
				ch <- response{name, nil, err}
				return
			}
			ch <- response{name, cmt, nil}
		}(tag, name, ch)
	}

	for _ = range cmts {
		r := <-ch
		if r.err != nil {
			fmt.Printf("**error** setup of [%s] failed: %v\n", r.name, r.err)
			return nil, r.err
		}
		cmts[r.name] = r.cmt
	}

	return diff, err
}
