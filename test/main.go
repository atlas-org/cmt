package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/atlas-org/cmt"
)

func main() {
	sh, err := cmt.NewShell()
	if err != nil {
		panic(err)
	}
	fmt.Printf(">>> starting shell...\n")
	go func() {
		for {
			io.Copy(os.Stdout, sh.Stdout)
		}
	}()

	err = sh.Setenv("TOTO", "1")
	time.Sleep(5 * time.Second)
	if err != nil {
		panic(err)
	}
	val := sh.Getenv("TOTO")
	time.Sleep(5 * time.Second)

	fmt.Fprintf(os.Stdout, "TOTO=%v\n", val)
}
