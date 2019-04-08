package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
	}

	client = &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}

	workerLimit = 5
)

type result struct {
	index  int
	count  int
	source string
}

func main() {
	getWordCountFromSources(os.Stdin, os.Stdout)
}

func getWordCountFromSources(input io.Reader, output io.Writer) {
	// Канал, ограничивающий количество одновременно запущенных горутин.
	qoutaC := make(chan struct{}, workerLimit)
	// Канал с результатами подсчетов.
	resultsC := make(chan *result, 1)

	wgResult := &sync.WaitGroup{}
	wgWorkers := &sync.WaitGroup{}

	// Запускаем в отдельной рутине функцию, для вывода результатов работы.
	go printResult(resultsC, output, wgResult)
	wgResult.Add(1)

	// Считываем Stdin.
	in := bufio.NewScanner(input)
	for i := 0; in.Scan(); i++ {
		source := in.Text()
		var content io.ReadCloser
		var err error

		if isURL(source) {
			content, err = getContentFromSite(source)
		} else {
			content, err = getContentFromFile(source)
		}
		if err != nil {
			continue
		}

		// Пробуем положить пустую структуру в канал.
		qoutaC <- struct{}{}
		wgWorkers.Add(1)

		r := &result{
			index:  i,
			source: source,
		}
		go wordCounter(content, r, resultsC, qoutaC, wgWorkers)
	}

	wgWorkers.Wait()

	close(qoutaC)
	close(resultsC)

	wgResult.Wait()
}

// printResult - вычитывает из канала и накапливает результаты работы подсчета.
// Сортирует результаты в порядке подачи источников.
// По окончанию работы программы выводит результаты и суммарное количество найденных совпадений.
func printResult(resultsC chan *result, output io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()

	var results []result
	var total int

	for val := range resultsC {
		results = append(results, *val)
		total += val.count
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	for _, result := range results {
		fmt.Fprintln(output, fmt.Sprintf("Count for %s: %d", result.source, result.count))
	}

	fmt.Fprintln(output, "Total:", total)
}

// getContentFromFile - получает контент из файла.
func getContentFromFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Println(fmt.Sprintf("Read file error: %v", err))
		return nil, err
	}

	return file, nil
}

// getContentFromSite - получает контент с сайта.
func getContentFromSite(url string) (io.ReadCloser, error) {
	resp, err := client.Get(url)
	if err != nil {
		log.Println(fmt.Sprintf("Get request error: %v", err))
		return nil, err
	}

	return resp.Body, nil
}

// wordCounter - считает количество вхождений строки в контенте.
func wordCounter(content io.ReadCloser, r *result, resultsC chan *result, qoutaC chan struct{}, wg *sync.WaitGroup) {
	defer content.Close()
	defer wg.Done()

	scanner := bufio.NewScanner(content)
	for i := 0; scanner.Scan(); i++ {
		r.count += strings.Count(scanner.Text(), "Go")
	}

	resultsC <- r
	<-qoutaC
}

// isURL - проверяет, является ли источник урлом.
func isURL(source string) bool {
	url, err := url.ParseRequestURI(source)
	if err != nil {
		return false
	}
	if len(url.Scheme) == 0 {
		return false
	}

	return true
}
