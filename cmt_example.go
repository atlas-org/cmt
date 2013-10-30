//+build ignore

package main

import (
	"fmt"

	"github.com/atlas-org/cmt"
)

const verbose = false

func main() {
	fmt.Printf("::: setting up a CMT environment...\n")
	mgr, err := cmt.NewMgr("rel1,devval", verbose)
	if err != nil {
		panic(err)
	}
	defer mgr.Delete()

}
