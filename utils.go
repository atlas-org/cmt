package cmt

import (
	//"fmt"
)

// project_release returns the name of the package holding
// the list of packages defining the release
func project_release(p *Project) string {
	n,ok := map[string]string{
		"LCGCMT": "LCG_Release",
		"dqm-common": "DQMCRelease",
		"tdaq-common": "TDAQCRelease",
	}[p.Name]
	if !ok {
		n = p.Name + "Release"
	}
	return n
}

// has_project returns true if p is in the slice
func has_project(projs []*Project, p *Project) bool {
		for _, pp := range projs {
			if pp == p {
				return true
			}
		}
		return false
}

// EOF
