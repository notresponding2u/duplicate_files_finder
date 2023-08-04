package main

import (
	"flag"
	"os"
)

func main() {
	hidden := flag.Bool("h", false, "Look into hidden files.")
	deleteOnDuplicate := flag.Bool("d", false, "Delete the duplicates.")
	silent := flag.Bool("s", false, "Silent mode.")
	flag.Parse()

	args := os.Args[1:]
	directory := ""

	if len(args) == 0 {
		directory = "./"
	} else {
		directory = args[len(args)-1]
	}

	p := New(func(name string) (handlerIface, error) {
		return os.Open(name)
	}, hidden, deleteOnDuplicate, silent)

	if err := p.Read(directory); err != nil {
		panic(err)
	}

	if err := p.Comparator(); err != nil {
		panic(err)
	}

	//for i := range p.GetDuplicates() {
	//	fmt.Printf("Duplicate: %s\n", p.GetDuplicates()[i].GetPath())
	//
	//}

	if err := p.Act(); err != nil {
		panic(err)
	}
}
