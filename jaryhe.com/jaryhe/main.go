package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"jaryhe.com/rsm"
)

type aa struct {
	x int
}

type bb struct {
	x int
}

type cc struct {
	x int
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test hly!\n"))
}

func (a aa) ServeHTTP(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	fmt.Printf("a = %d\n", a.x)
	fmt.Println(m)
	w.Write([]byte("test aaaaaaaaaaaaa!\n"))
}

func (b bb) ServeHTTP(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	fmt.Printf("b = %d\n", b.x)
	fmt.Println(m)
	w.Write([]byte("test bbbbbbbbbbbbb!\n"))
}

func (c cc) ServeHTTP(w http.ResponseWriter, r *http.Request, m map[string]interface{}) {
	fmt.Printf("c = %d\n", c.x)
	fmt.Println(m)
	w.Write([]byte("test ccccccccccccc!\n"))
}

func main() {
	serveMux := rsm.NewServeMux()
	a := aa{x: 1}
	b := bb{x: 2}
	c := cc{x: 3}

	/*	var hander io.Writer
		hander = a
		fmt.Println(hander.(value))*/

	serveMux.Handle("/aa/:dfadf/vm/:vmid", a)
	serveMux.Handle("/bb/:name", b)
	serveMux.Handle("/cc/:id", c)
	s := &http.Server{
		Addr:           ":8080",
		Handler:        serveMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
