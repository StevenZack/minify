package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/StevenZack/openurl"
)

var open = flag.Bool("open", false, "open in browser after build")

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
	const out = "docs"

	// dependencies
	checkNpmDeps("html-minifier", "")
	checkNpmDeps("uglifyjs", "uglify-js")
	checkNpmDeps("css-minify", "")

	// remove dist/*
	os.RemoveAll(out)

	// template
	var root *template.Template
	e = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		path = filepath.ToSlash(path)

		switch filepath.Ext(info.Name()) {
		case ".html":
			relativeUri := strings.TrimPrefix(strings.TrimPrefix(path, dir), "/") // like /index.html
			if root == nil {
				root = template.New(relativeUri)
			}

			var t *template.Template
			if relativeUri == root.Name() {
				t = root
			} else {
				t = root.New(relativeUri)
			}

			//read
			fi, e := os.OpenFile(path, os.O_RDONLY, 0644)
			if e != nil {
				return e
			}
			defer fi.Close()
			b, e := io.ReadAll(fi)
			if e != nil {
				return e
			}

			_, e = t.Parse(string(b))
			if e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		log.Println(e)
		return
	}

	//walk
	e = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		path = filepath.ToSlash(path)
		relativePath := strings.TrimPrefix(strings.TrimPrefix(path, dir), "/")
		if strings.HasPrefix(relativePath, out) {
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
		if root != nil && strings.HasSuffix(relativePath, ".html") {
			fo, e := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
			if e != nil {
				log.Println(e)
				return e
			}
			defer fo.Close()
			e = root.ExecuteTemplate(fo, relativePath, nil)
			if e != nil {
				log.Println(e)
				return e
			}

		} else {
			e = copyFile(path, dst)
			if e != nil {
				log.Println(e)
				return e
			}
		}

		e = tryMinify(dst)
		if e != nil {
			log.Println(e)
			return e
		}
		return nil
	})

	if *open {
		port := RandomPort()
		http.Handle("/", &Router{dir: out})
		addr := "http://localhost:" + strconv.Itoa(port)
		fmt.Println(addr)
		openurl.Open(addr)
		e = http.ListenAndServe(":"+strconv.Itoa(port), nil)
		if e != nil {
			log.Println(e)
			return
		}
	}
}

func RandomPort() int {
	return rand.Intn(10000) + 10000
}

func copyFile(path, dst string) error {
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
