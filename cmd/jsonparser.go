package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Avik32223/jsonparser/internal/jsonparser"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		panic("a JSON file must be passed as arguments. None provided")
	}
	f, err := os.ReadFile(args[0])
	if err != nil {
		panic(fmt.Sprintf("error reading file %s", args[0]))
	}
	p := jsonparser.NewParser(string(f))
	m, err := p.Parse()
	if err != nil {
		fmt.Println(m)
		os.Exit(1)
	}
	fmt.Printf("%#v\n", m)
	os.Exit(0)
}
