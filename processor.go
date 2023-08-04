package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"
)

type processor struct {
	entries          []*entry
	duplicates       []*entry
	hidden           *bool
	deleteDuplicates *bool
	silent           *bool
	list             *bool
	opener           func(name string) (handlerIface, error)
}

type handlerIface interface {
	Readdir(n int) ([]fs.FileInfo, error)
	Close() error
}

type entry struct {
	index int
	fs.FileInfo
	fullPath string
}

func (e *entry) SetIndex(index int) {
	e.index = index
}

func (e *entry) GetIndex() int {
	return e.index
}

func (e *entry) GetPath() string {
	return e.fullPath
}

func New(open func(name string) (handlerIface, error), hidden *bool, deleteOnDuplicate *bool, silent *bool) *processor {

	return &processor{
		hidden:           hidden,
		deleteDuplicates: deleteOnDuplicate,
		silent:           silent,
		opener:           open,
	}
}

func (p *processor) GetEntries() []*entry {
	return p.entries
}

func (p *processor) GetDuplicates() []*entry {
	return p.duplicates
}

func (p *processor) Read(directory string) error {
	f, err := p.opener(directory)
	if err != nil {
		return err
	}

	defer f.Close()

	elements, err := f.Readdir(0)
	if err != nil {
		return err
	}

	for i := range elements {
		if elements[i].Name()[0] == '.' && !*p.hidden {
			continue
		}

		pth := ""
		if directory[len(directory)-1] == '/' {
			pth = fmt.Sprintf("%s%s", directory, elements[i].Name())
		} else {
			pth = fmt.Sprintf("%s/%s", directory, elements[i].Name())
		}

		if elements[i].IsDir() {
			err := p.Read(pth)
			if err != nil {
				return err
			}

			continue
		}

		p.entries = append(p.entries, &entry{
			FileInfo: elements[i],
			fullPath: pth,
		})
	}

	return nil
}

func (p *processor) Comparator() error {
	thread := 64
	queue := make(chan *entry, thread)
	g := errgroup.Group{}

	for i := 0; i < thread; i++ {
		g.Go(func() error {
			for e := range queue {
				err := p.compare(e)
				if err != nil {
					return err
				}
			}

			return nil
		})
	}

	for i := range p.GetEntries() {
		p.GetEntries()[i].SetIndex(i)

		queue <- p.GetEntries()[i]
	}

	close(queue)

	return g.Wait()
}

func (p *processor) compare(e *entry) error {
	for i := range p.GetEntries() {
		if i >= e.GetIndex() {
			return nil
		}

		if e.Name() == p.GetEntries()[i].Name() {
			p.duplicates = append(p.duplicates, e)

			return nil
		}
	}

	return nil
}

func (p *processor) Act() error {
	for i := range p.GetDuplicates() {
		var d *bool

		if (p.silent == nil || !*p.silent) && (p.list == nil || !*p.list) {
			d = p.confirm(p.GetDuplicates()[i].GetPath())
		}

		if (p.silent != nil && *p.silent) || (d != nil && *d) {
			if err := os.Remove(p.GetDuplicates()[i].GetPath()); err != nil {
				return err
			}
		} else if p.list != nil && *p.list {
			fmt.Printf("Duplicate: %s\n", p.GetDuplicates()[i].GetPath())
		}
	}

	return nil
}

func (p *processor) confirm(s string) *bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Do you want to delete file %s\n", s)
		fmt.Print(" y - yes\n n - no\n a - all\n l - list only\n")
		fmt.Print("[y/n/a/l]: ")

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "y", "yes":
			return pointer(true)
		case "n", "no":
			return pointer(false)
		case "a", "all":
			p.silent = pointer(true)
			return pointer(true)
		case "l", "list":
			p.list = pointer(true)
			return pointer(false)
		}
	}

	return nil
}
func pointer[T any](e T) *T {
	return &e
}
