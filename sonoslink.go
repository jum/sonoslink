package main

import (
	"bufio"
	"flag"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
)

var (
	iTunesRoot = flag.String("src", "/Volumes/Hydra/Musik", "iTunes original music folder")
	sonosRoot  = flag.String("dst", "/Volumes/Hydra/SonosMusik", "Sonos music folder")
	srcList    = flag.String("list", "", "file name with list of subdirs")
	//iTunesRoot = flag.String("src", "test", "iTunes original music folder")
	//sonosRoot  = flag.String("dst", "sonos", "Sonos music folder")
)

const DEBUG = false

func debug(format string, a ...interface{}) {
	if DEBUG {
		fmt.Printf(format, a...)
	}
}

func main() {
	flag.Parse()
	debug("src %v, dst %v\n", *iTunesRoot, *sonosRoot)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			debug("walkFn path %v, err %v\n", path, err)
			return err
		}
		// continue to walk all directories
		if info.IsDir() {
			return nil
		}
		// we are only concerned with music files
		ext := filepath.Ext(path)
		if !(ext == ".mp3" || ext == ".m4a") {
			debug("skipping path %v\n", path)
			return nil
		}
		debug("doing path %v, %#v\n", path, info)
		dir, file := filepath.Split(path)
		dirHash := hashString(dir)
		fileHash := hashString(file)
		debug("dir %v(%v), file %v(%v)\n", dir, dirHash, file, fileHash)
		newDir := filepath.Join(*sonosRoot, dirHash)
		debug("new file %v/%v\n", newDir, fileHash)
		err = os.MkdirAll(newDir, 0755)
		if err != nil {
			panic(err)
		}
		err = os.Link(path, filepath.Join(newDir, fileHash)+ext)
		if err != nil {
			panic(err)
		}
		return nil
	}
	if flag.NArg() > 0 {
		for _, dname := range flag.Args() {
			debug("arg %v\n", dname)
			err := filepath.Walk(dname, walkFunc)
			if err != nil {
				panic(err)
			}
		}
	} else {
		if len(*srcList) > 0 {
			f, err := os.Open(*srcList)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			s := bufio.NewScanner(f)
			for s.Scan() {
				l := s.Text()
				debug("%v\n", l)
				err := filepath.Walk(filepath.Join(*iTunesRoot, l), walkFunc)
				if err != nil {
					panic(err)
				}
			}
			err = s.Err()
			if err != nil {
				panic(err)
			}
		} else {
			if len(*iTunesRoot) > 0 {
				err := filepath.Walk(*iTunesRoot, walkFunc)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func hashString(s string) string {
	h := crc32.NewIEEE()
	_, err := h.Write([]byte(s))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%04x", h.Sum32())
}
