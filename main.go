package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.Usage = func() {
		fmt.Println("minify [-options]")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	dir := "."
	var e error
	dir, e = filepath.Abs(dir)
	if e != nil {
		log.Println(e)
		return
	}
	dir = filepath.ToSlash(dir)
	const out = "dist"

	// dependencies
	checkNpmDeps("html-minifier", "")
	checkNpmDeps("uglifyjs", "uglify-js")
	checkNpmDeps("css-minify", "")

	// remove dist/*
	os.RemoveAll(out)

	//walk
	e = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		path = filepath.ToSlash(path)
		relativePath := strings.TrimPrefix(path, dir)
		if strings.HasPrefix(strings.TrimPrefix(relativePath, "/"), out) {
			// dist/*
			return nil
		}
		dst := filepath.Join(out, relativePath)

		e := os.MkdirAll(filepath.Dir(dst), 0744)
		if e != nil {
			log.Println(e)
			return e
		}

		fmt.Println(relativePath)
		e = CopyFile(path, dst)
		if e != nil {
			log.Println(e)
			return e
		}

		e = tryMinify(dst)
		if e != nil {
			log.Println(e)
			return e
		}
		return nil
	})
}

func CopyFile(path, dst string) error {
	b, e := ioutil.ReadFile(path)
	if e != nil {
		log.Println(e)
		return e
	}
	e = ioutil.WriteFile(dst, b, 0644)
	if e != nil {
		log.Println(e)
		return e
	}
	return nil
}

func checkNpmDeps(cmd, pkg string) {
	if pkg == "" {
		pkg = cmd
	}

	e := exec.Command(cmd, "-h").Run()
	if e != nil {
		log.Println(e)
		log.Fatal("Missing dependency: sudo npm i -g " + pkg)
	}
}

func tryMinify(fi string) error {
	dir := filepath.Dir(fi)
	switch filepath.Ext(fi) {
	case ".css":
		e := exec.Command("css-minify", "-f", fi, "-o", dir).Run()
		if e != nil {
			log.Println(e)
			return e
		}
		e = os.Rename(filepath.Join(dir, minifiedName(fi)), fi)
		if e != nil {
			log.Println(e)
			return e
		}
		return nil
	case ".js":
		e := exec.Command("uglifyjs", "--compress", "--mangle", "-o", fi, fi).Run()
		if e != nil {
			log.Println(e)
			return e
		}
		return nil
	case ".html":
		e := exec.Command("html-minifier", "--caseSensitive", "--collapse-whitespace", "--remove-comments", "--remove-optional-tags", "--remove-redundant-attributes", "--remove-script-type-attributes", "--remove-tag-whitespace", "--minify-css", "true", "--minify-js", "true", "-o", fi, fi).Run()
		if e != nil {
			log.Println(e)
			return e
		}
		return nil
	default:
		return nil
	}
}

func minifiedName(fi string) string {
	base := filepath.Base(fi)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)] + ".min" + ext
}
