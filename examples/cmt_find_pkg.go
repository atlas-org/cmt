//+build ignore

package main

import (
	"fmt"

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

	pkg := "Control/AthenaKernel"
	fmt.Printf("checkout [%s]...\n", pkg)
	err = cmt.CheckOut(pkg, "")
	if err != nil {
		panic(err)
	}

	vers := cmt.PackageVersion("Control/AthenaServices")
	fmt.Printf("==> version =%q\n", vers)

	p, err := cmt.Package("AthenaServices")
	if err != nil {
		panic(err)
	}
	fmt.Printf("==> %v\n", *p)
}

// EOF
