package search

import (
	"context"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

//Describe searches result

type Result struct {
	Phrase  string
	Line    string
	LineNum int64
	ColNum  int64
}

//All - total searching
func All(ctx context.Context, phrase string, files []string) <-chan []Result {
	ch := make(chan []Result)
	wg := sync.WaitGroup{}
	//root := context.Background()
	ctx, cansel := context.WithCancel(ctx)

	for i := 0; i < len(files); i++ {
		wg.Add(1)

		go func(ctx context.Context, currentFile string, i int, ch chan<- []Result) {
			defer wg.Done()
			result, err := findAll(currentFile, phrase)

			if err != nil {
				log.Println(err)
			}

			if len(result) > 0 {
				ch <- result
			}

		}(ctx, files[i], i, ch)
	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	cansel()
	return ch
}

func findAll(filePath string, phrase string) (res []Result, err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("ERROR, file not found")
		return res, err
	}
	arr := strings.Split(string(data), "\n")

	for i, str := range arr {
		ind := strings.Index(str, phrase)
		if ind > -1 {
			found := Result{
				Phrase:  phrase,
				Line:    str,
				LineNum: int64(i + 1),
				ColNum:  int64(ind) + 1,
			}
			res = append(res, found)
		}
	}
	return res, nil
}

//Any - search first phrase
func Any(ctx context.Context, phrase string, files []string) <-chan Result {
	ch := make(chan Result)
	wg := sync.WaitGroup{}
	result := Result{}

	for i := 0; i < len(files); i++ {
		data, err := ioutil.ReadFile(files[i])
		if err != nil {
			log.Println("error opening file: ", err)
		}

		if strings.Contains(string(data), phrase) {
			res := FindAny(phrase, string(data))
			if (Result{}) != res {
				result = res
				break
			}
		}
	}

	wg.Add(1)
	go func(ctx context.Context, ch chan<- Result) {
		defer wg.Done()
		if (Result{}) != result {
			ch <- result
		}
	}(ctx, ch)

	go func() {
		defer close(ch)
		wg.Wait()
	}()
	return ch
}

func FindAny(phrase, search string) (result Result) {
	for i, line := range strings.Split(search, "\n") {
		if strings.Contains(line, phrase) {
			return Result{
				Phrase:  phrase,
				Line:    line,
				LineNum: int64(i + 1),
				ColNum:  int64(strings.Index(line, phrase)) + 1,
			}
		}
	}
	return result
}
