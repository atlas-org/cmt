//+build ignore

package main

import (
	"fmt"
	"os"

	gocmt "github.com/atlas-org/cmt"
)

const verbose = true

func main() {
	fmt.Printf("::: setting up a CMT environment...\n")
	setup, err := gocmt.NewSetup("rel1,devval", verbose)
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

	fmt.Printf("::: storing CMT environment into [store.cmt]...\n")
	f, err := os.Create("store.cmt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = setup.Save(f)
	if err != nil {
		panic(err)
	}

}

// EOF
