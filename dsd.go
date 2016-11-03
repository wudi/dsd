package main

import (
	"flag"
	"strings"
	"time"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"sync"
)

type DriverFactory interface {
	Search(word string) bool
}

type Gandi struct {
	Endpoint string
	Retry    int
}

type GandiResp struct {
	Available string `json:"available"`
}

func NewGandi() *Gandi {
	return &Gandi{
		Endpoint: "https://www.gandi.net/domain/suggest/verbose_tlds?currency=CNY&tld=%s",
		Retry:5,
	}
}

func (g *Gandi) Search(word string) (r bool) {
	fmt.Printf("Driver[Gandi]: search word `%s`\n", word)

	endpoint := fmt.Sprintf(g.Endpoint, word)
	request, _ := http.NewRequest(http.MethodGet, endpoint, nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.87 Safari/537.36")
	request.Header.Add("Referer", "https://www.gandi.net/domain/suggest")
	request.Header.Add("Accept", "application/json, text/plain, */*")

	client := &http.Client{
		Timeout:timeout,
	}

	t := 0
	RETRY:
	if t > g.Retry {
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("Driver[Gandi]: http request, %s\n", err.Error())
		return
	}

	contents, _ := ioutil.ReadAll(resp.Body)
	if debug {
		fmt.Printf("%s\n%s\n", endpoint, string(contents))
	}

	var v []*GandiResp
	if err := json.Unmarshal(contents, &v); err != nil {
		fmt.Printf("Driver[Gandi]: json Unmarshal, %s\n", err.Error())
		return
	}

	if len(v) > 0 {
		if v[0].Available == "available" {
			r = true
		} else if v[0].Available == "pending" {
			t++
			goto RETRY
		}
	}

	return
}

var words string = ""
var concurrentNum int = 5
var sleep time.Duration = 1 * time.Second
var extensions string = ".com .cn .net"
var timeout time.Duration = 3 * time.Second
var debug bool = false

func init() {
	flag.StringVar(&words, "w", words, "Words eg: apple mac (split with single space)")
	flag.IntVar(&concurrentNum, "c", concurrentNum, "concurrent numbers")
	flag.DurationVar(&sleep, "s", sleep, "sleep seconds")
	flag.DurationVar(&timeout, "t", timeout, "timeout")
	flag.StringVar(&extensions, "e", extensions, "domain extensions, eg: .com .io .net")
	flag.BoolVar(&debug, "d", debug, "debug")
	flag.Parse()
}

func main() {
	w := strings.Split(words, " ")
	exts := strings.Split(extensions, " ")
	gandi := NewGandi()

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
			go func(ext string) {
				if gandi.Search(ext) {
					fmt.Printf("%s: \033[;32mavailable\033[0m\n", ext)
				} else {
					fmt.Printf("%s: unavailable\n", ext)
				}

				wg.Done()
			}(v + e)
		}

		if (i > 0) && (i % concurrentNum == 0) {
			time.Sleep(sleep)
		}
	}

	wg.Wait()
}
