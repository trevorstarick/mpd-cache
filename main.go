package main

import (
	"io"
	"os"
	"path"

	"github.com/fhs/gompd/mpd"
)

var active = make(map[string]struct{})

const srcDir = "/azure/data/music"
const destDir = "/home/trevorstarick/Music"

func cacheSong(filename string) {
	if _, exists := active[filename]; exists {
		return
	}

	active[filename] = struct{}{}

	sourceFilename := srcDir + "/" + filename
	finalFilename := destDir + "/" + filename
	tempFilename := finalFilename + ".temp"

	s, err := os.Lstat(srcDir)
	if err != nil {
		panic(err)
	}
	isLocal := s.Mode()&os.ModeSymlink == os.ModeSymlink

	if isLocal {
		return
	}

	if _, err := os.Stat(finalFilename); os.IsNotExist(err) {
		dir := path.Dir(finalFilename)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			panic(err)
		}

		from, err := os.Open(sourceFilename)
		if err != nil {
			panic(err)
		}
		defer from.Close()

		to, err := os.OpenFile(tempFilename, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			panic(err)
		}

		err = os.Rename(tempFilename, finalFilename)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	var previousSong string
	conn, err := mpd.DialAuthenticated("tcp", ":6600", "")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		song, _ := conn.CurrentSong()

		if song["file"] != previousSong {
			go cacheSong(song["file"])
			previousSong = song["file"]
		}
	}
}
