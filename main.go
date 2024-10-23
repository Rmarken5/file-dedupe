package main

import (
	"flag"
	"github.com/rmarken5/file-dedupe/hasher"
	"log"
)

var (
	directory = flag.String("d", ".", "directory to search")
)

func main() {
	flag.Parse()

	m := hasher.NewManager()
	_, err := m.Run(*directory)
	if err != nil {
		log.Fatal(err)
	}

}
