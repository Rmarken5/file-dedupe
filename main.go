package main

import (
	"context"
	"flag"
	"github.com/rmarken5/file-dedupe/hasher"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	directory = flag.String("d", ".", "directory to search")
)

func main() {
	// Create a CPU profile file
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}

	defer func() {
		if tmpErr := f.Close(); tmpErr != nil {
			err = tmpErr
		}
	}()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	flag.Parse()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	m := hasher.NewManager()
	_, err = m.Run(context.Background(), *directory)
	if err != nil {
		log.Fatal(err)
	}

	me, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer me.Close() // error handling omitted for example
	runtime.GC()     // get up-to-date statistics
	if err := pprof.WriteHeapProfile(me); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

}
