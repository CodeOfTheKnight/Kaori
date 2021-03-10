package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func ls() {
	var files []string

	root := "/gui"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
	}
}
