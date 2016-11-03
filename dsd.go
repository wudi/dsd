package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var words string = ""
var concurrentNum int = 5
var sleep time.Duration = 1 * time.Second
var extensions string = ".com .cn .net"
var timeout time.Duration = 3 * time.Second
var debug bool = false
var onlyPrintAvailable = false

func init() {
	flag.StringVar(&words, "w", words, "Words eg: apple mac (split with single space)")
	flag.IntVar(&concurrentNum, "c", concurrentNum, "concurrent numbers")
	flag.DurationVar(&sleep, "s", sleep, "sleep seconds")
	flag.DurationVar(&timeout, "t", timeout, "timeout")
	flag.StringVar(&extensions, "e", extensions, "domain extensions, eg: .com .io .net")
	flag.BoolVar(&debug, "d", debug, "debug")
	flag.BoolVar(&onlyPrintAvailable, "a", onlyPrintAvailable, "only print available domains")
	flag.Parse()
}

func main() {
	w := strings.Split(words, " ")
	exts := strings.Split(extensions, " ")
	if len(w) == 0 {
		fmt.Println("Please input the domain word.")
		return
	}

	if len(exts) == 0 {
		fmt.Println("Invalid domain extensions.")
		return
	}

	var wg sync.WaitGroup
	for i, v := range w {
		if len(v) == 0 || v == " " {
			continue
		}

		wg.Add(len(exts))
		for _, e := range exts {
			go func(domain string) {
				ns, err := net.LookupNS(domain)
				if err != nil {
					fmt.Printf("%s: \033[;32mavailable\033[0m\n", domain)
				} else {
					fmt.Printf("%s: unavailable\n", domain)
				}

				if debug {
					s := make([]string, len(ns))
					for i, v := range ns {
						s[i] = v.Host
					}
					fmt.Printf("LookupNS[%s]: %s\n", s)
				}
				wg.Done()
			}(v + e)
		}

		if (i > 0) && (i%concurrentNum == 0) {
			time.Sleep(sleep)
		}
	}

	wg.Wait()
}
