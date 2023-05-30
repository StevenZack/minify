package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Router struct {
	dir string
}

func (s *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if p == "/" {
		p = "/index.html"
	}
	path := filepath.Join(s.dir, p)
	info, e := os.Stat(path)
	if e != nil {
		if os.IsNotExist(e) {
			http.ServeFile(w, r, filepath.Join(s.dir, "404.html"))
			return
		}
		log.Println(e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.ServeFile(w, r, filepath.Join(path, "index.html"))
		return
	}
	http.ServeFile(w, r, path)
}
