package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

func main() {
	src := flag.String("src", "", "source folder")
	dst := flag.String("dst", "", "destination folder")
	verbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()

	if len(*src) == 0 {
		flag.Usage()
		log.Fatal("invalid source folder")
	}
	if len(*dst) == 0 {
		flag.Usage()
		log.Fatal("invalid destination folder")
	}

	if _, err := os.Stat(*src); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", *src)
	}

	srcChan := make(chan string)
	go processFiles(srcChan, *dst, *verbose)

	filepath.Walk(*src, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(strings.ToLower(path), ".jpg") {
			srcChan <- path
		}
		return nil
	})

	close(srcChan)
}

func processFiles(srcChan <-chan string, dst string, verbose bool) {
	for file := range srcChan {
		if verbose {
			log.Println(file)
		}
		f, err := os.Open(file)
		if err != nil {
			log.Printf("open: %v (skipping): %s", err, file)
			continue
		}

		x, err := exif.Decode(f)
		if err != nil {
			log.Printf("decode: %v (skipping): %s", err, file)
			continue
		}

		tm, err := x.DateTime()
		if err != nil {
			log.Printf("datetime: %v (skipping): %s", err, file)
			continue
		}
		if verbose {
			log.Println("Taken: ", tm)
		}
		f.Close()
		dstPath := path.Join(dst, strconv.Itoa(tm.Year()), strconv.Itoa(int(tm.Month())))
		os.MkdirAll(dstPath, os.ModePerm)
		_, dstFile := filepath.Split(file)
		dstPath = path.Join(dstPath, dstFile)
		if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
			log.Println("file exists:", dstPath, "(skipping):", file)
			continue
		}
		if verbose {
			log.Println("copying to", dstPath)
		}
		copy(dstPath, file)
	}
}

func copy(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}
