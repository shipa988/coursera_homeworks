package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"os"
)

func tern(condition bool, truvalue interface{}, falsevalue interface{}) interface{} {
	if condition {
		return truvalue
	} else {
		return falsevalue
	}
}
func PrintStruct(out io.Writer, fileinfo os.FileInfo, islast bool, prefix string /*,predokislast bool,level int*/) error {
	_, err := fmt.Fprint(out,
		prefix,
		tern(islast, "└───", "├───"),
		fileinfo.Name(),
		tern(fileinfo.IsDir(), "", " ("+tern(fileinfo.Size()==0,"empty", strconv.Itoa(int(fileinfo.Size()))+"b").(string)+")"),
		"\n")
	return err
}

func LeveldirTree(out io.Writer, path string, printingFiles bool, prefix string, /*level int,predokislast bool*/) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("Path Could not be read: %v", err)
	}
	lastdir := ""
	for i := range files {
		if files[i].IsDir() {
			lastdir = files[i].Name()
		}
	}

	for i := range files {
		islastDir := files[i].Name() == lastdir
		islastFile := i == (len(files) - 1)
		if printingFiles {
			PrintStruct(out, files[i], islastFile, prefix)
			if files[i].IsDir() {
				LeveldirTree(out, filepath.Join(path, files[i].Name()), printingFiles, tern(islastFile, prefix+"	", prefix+"│	").(string))
			}
		} else {
			if files[i].IsDir() {
				PrintStruct(out, files[i], islastDir, prefix)
				LeveldirTree(out, filepath.Join(path, files[i].Name()), printingFiles, tern(islastDir, prefix+"	", prefix+"│	").(string))
			}
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printingFiles bool) error {
	var prefix string = ""
	err := LeveldirTree(out, path, printingFiles, prefix)
	return err
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	//err := dirTree(out, "test/testdata", false)
	if err != nil {
		panic(err.Error())
	}
}
