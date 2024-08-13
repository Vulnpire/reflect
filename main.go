package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)

	jobs := make(chan string)
	var wg sync.WaitGroup

	keywords := []string{
		"alert(",
		"confirm(",
		"prompt(",
	}

	isValidXSS := func(domain, body string) bool {
		u, err := url.Parse(domain)
		if err != nil {
			return false
		}

		for _, keyword := range keywords {
			if strings.Contains(body, keyword) {
				if strings.Contains(body, "<script>") || strings.Contains(body, "onerror=") || strings.Contains(body, u.Query().Encode()) {
					return true
				}
			}
		}
		return false
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range jobs {
				resp, err := http.Get(domain)
				if err != nil {
					continue
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}
				sb := string(body)

				if isValidXSS(domain, sb) {
					fmt.Println("Reflected:", domain)
				}
			}
		}()
	}

	for sc.Scan() {
		domain := sc.Text()
		jobs <- domain
	}

	close(jobs)
	wg.Wait()
}
