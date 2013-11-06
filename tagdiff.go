package cmt

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// TagDiff returns the list of tag differences between 2 releases/nightlies
func TagDiff(old, new string, display, verbose bool) (map[string]map[string]Package, error) {
	var err error
	diffs := make(map[string]map[string]Package)

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
		cmt  *Cmt
		err  error
	}

	ch := make(chan response)
	defer close(ch)

	for name, tag := range tags {
		go func(tag, name string, ch chan response) {
			if display {
				fmt.Printf("::: setup %s env. [%s]...\n", name, tag)
			}
			env, err := NewSetup(tag, verbose)
			if err != nil {
				env.Delete()
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

	pkgs := map[string]map[string]Package{
		"old": make(map[string]Package),
		"new": make(map[string]Package),
	}

	for k := range cmts {
		cmt := cmts[k]
		dag, err := cmt.ProjectsDag()
		if err != nil {
			return nil, err
		}
		for _, proj := range dag {
			dirnames, err := filepath.Glob(filepath.Join(proj.Path, "*Release"))
			if err != nil {
				return diffs, err
			}
			if len(dirnames) != 1 {
				continue
			}
			reldata := filepath.Join(dirnames[0], "cmt", "requirements")
			uses, err := extract_uses(reldata, cmt.msg)
			if err != nil {
				return diffs, err
			}
			projname := proj.Name
			if strings.HasPrefix(projname, "Atlas") {
				projname = projname[len("Atlas"):]
			}
			for _, use := range uses {
				use.Project = projname
				pkgs[k][use.Name] = use
			}
		}
	}

	cmp_diffs := func(a, b string) {
		for pname := range pkgs[a] {
			p_a := pkgs[a][pname]
			p_b, ok := pkgs[b][pname]
			if !ok {
				diffs[pname] = map[string]Package{
					a: p_a,
					b: Package{"None", "None-00-00-00", p_a.Project},
				}
				continue
			}
			if p_b.Version != p_a.Version {
				diffs[pname] = map[string]Package{
					a: p_a,
					b: p_b,
				}
				continue
			}
		}

	}
	cmp_diffs("old", "new")
	cmp_diffs("new", "old")

	if len(diffs) == 0 {
		return nil, nil
	}

	if !display {
		return diffs, err
	}

	format := "%-15s %-15s | %-15s %-15s | %-45s\n"
	fmt.Printf(format, "old", "old-proj", "new", "new-project", "pkg-name")
	fmt.Printf(strings.Repeat("-", 120) + "\n")

	keys := make([]string, 0, len(diffs))
	for pname, _ := range diffs {
		keys = append(keys, pname)
	}
	sort.Strings(keys)

	for _, pname := range keys {
		diff := diffs[pname]
		p_old := diff["old"]
		p_new := diff["new"]
		v_old := strings.Replace(p_old.Version, p_old.Base()+"-", "", -1)
		v_new := strings.Replace(p_new.Version, p_new.Base()+"-", "", -1)

		fmt.Printf(format,
			v_old, p_old.Project,
			v_new, p_new.Project,
			pname,
		)
	}
	fmt.Printf(strings.Repeat("-", 120) + "\n")
	fmt.Printf("::: found [%d] tags which are different\n", len(diffs))

	return diffs, err
}
