package main

import (
	"fmt"
	//"io"
	"os"
	//"time"

	"github.com/atlas-org/cmt"
)

func main() {
	sh, err := cmt.NewShell()
	if err != nil {
		panic(err)
	}
	fmt.Printf(">>> starting shell...\n")

	err = sh.Setenv("TOTO", "101")
	if err != nil {
		panic(err)
	}
	err = sh.Setenv("TOTO", "1011")
	if err != nil {
		panic(err)
	}
	err = sh.Setenv("TOTO", "1012")
	if err != nil {
		panic(err)
	}
	{
		val := sh.Getenv("TOTO")
		fmt.Fprintf(os.Stdout, "TOTO=%q\n", val)
	}
	err = sh.Setenv("TOTO", "1011")
	if err != nil {
		panic(err)
	}
	{
		val := sh.Getenv("TATA")
		fmt.Fprintf(os.Stdout, "TATA=%q\n", val)
	}
	{
		val := sh.Getenv("TOTO")
		fmt.Fprintf(os.Stdout, "TOTO=%q\n", val)
	}
	err = sh.Source("toto.sh")
	if err != nil {
		panic(err)
	}
	{
		val := sh.Getenv("TITI")
		fmt.Fprintf(os.Stdout, "TITI=%q\n", val)
	}
}
