//+build ignore

package main

import (
	"fmt"
	"os"

	gocmt "github.com/atlas-org/cmt"
)

const verbose = true

func main() {
	fmt.Printf("::: loading up a CMT environment...\n")
	testarea, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	setup, err := gocmt.NewSetupFromCache("store.cmt", testarea, verbose)
	if err != nil {
		panic(err)
	}
	defer setup.Delete()

	cmt, err := gocmt.New(setup)
	if err != nil {
		panic(err)
	}

	projs, err := cmt.Projects()
	if err != nil {
		panic(err)
	}
	fmt.Printf("projects:\n")
	for _, p := range projs {
		fmt.Printf("%v\n", p)
	}

}

// EOF
