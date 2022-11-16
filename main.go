package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var url = "https://agritrop.cirad.fr/584726/1/Rapport.pdf"

func main() {
	fmt.Println("concurrency form:")

	t := time.Now()

	fmt.Println("START")

	err := ConcurrencyDownload(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("FINISH")

	fmt.Println(time.Since(t))

	os.Remove("Rapport.pdf")

	// ---

	fmt.Println("simple form:")

	t = time.Now()

	fmt.Println("START")

	err = SimpleDownload(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("FINISH")

	fmt.Println(time.Since(t))

	os.Remove("Rapport.pdf")
}

func ConcurrencyDownload(url string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("invalid url")
	}
	client := http.Client{}

	res, err := http.Head(url)
	if err != nil {
		return err
	}

	// verify if suport requset for file downloads
	if res.Header.Get("Accept-Ranges") != "bytes" {
		return errors.New("unable to download file with multithreads")
	}

	urlSplit := strings.Split(url, "/")

	fileName := urlSplit[len(urlSplit)-1]

	// size of file
	cntLen, err := strconv.Atoi(res.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	nbPart := 3
	offset := cntLen / nbPart

	// waitgroup for goroutine
	wg := sync.WaitGroup{}

	for i := 0; i < nbPart; i++ {
		wg.Add(1)

		name := fmt.Sprintf("part%d", i)
		start := i * offset
		end := (i + 1) * offset

		go func() {
			defer wg.Done()

			part, err := os.Create(name)
			if err != nil {
				return
			}
			defer part.Close()

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return
			}

			// Add range with start and end
			req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

			res, err := client.Do(req)
			if err != nil {
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				return
			}

			_, err = part.Write(body)
			if err != nil {
				return
			}
		}()
	}

	wg.Wait()

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	for i := 0; i < nbPart; i++ {
		name := fmt.Sprintf("part%d", i)

		file, err := ioutil.ReadFile(name)
		if err != nil {
			return err
		}

		out.WriteAt(file, int64(i*offset))

		if err := os.Remove(name); err != nil {
			return err
		}
	}

	return nil
}

func SimpleDownload(url string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("invalid url")
	}
	client := http.Client{}

	res, err := http.Head(url)
	if err != nil {
		return err
	}

	// verify if suport requset for file downloads
	if res.Header.Get("Accept-Ranges") != "bytes" {
		return errors.New("unable to download file with multithreads")
	}

	urlSplit := strings.Split(url, "/")

	fileName := urlSplit[len(urlSplit)-1]

	// do work

	func() {
		file, err := os.Create(fileName)
		if err != nil {
			return
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return
		}

		res, err := client.Do(req)
		if err != nil {
			return
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return
		}

		_, err = file.Write(body)
		if err != nil {
			return
		}
	}()
	return nil
}
