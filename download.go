package download

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

var mb int64 = 1024 * 1024

func Download(url string, filename string, goroutine int) error {
	contentLength, err := getContentLength(url)
	if err != nil {
		return err
	}

	last := lastFilename(filename)

	var parts []part
	var i int64 = 0
	for ; (i+1)*mb-1 < contentLength; i++ {
		parts = append(parts, part{index: int(i), url: url, filename: last, rangeL: mb * i, rangeR: mb*(i+1) - 1})
	}

	files, err := partDownload(parts, goroutine)
	if err != nil {
		return err
	}

	err = joinFile(files, filename)
	return err
}

func getContentLength(url string) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	return resp.ContentLength, nil
}

func download(url, file string, rangeL, rangeR int64) (string, error) {
	filename, err := ioutil.TempFile("", file+"-*")
	if err != nil {
		return "", err
	}
	defer filename.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rangeL, rangeR))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(filename, resp.Body); err != nil {
		return "", err
	}

	return filename.Name(), nil
}

type part struct {
	index    int
	url      string
	filename string
	rangeL   int64
	rangeR   int64
}

type result struct {
	index    int
	filename string
	err      error
}

func partDownload(parts []part, goroutine int) ([]string, error) {
	results := make(chan result, len(parts))
	p := make(chan part, len(parts))
	for _, v := range parts {
		p <- v
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < goroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case part := <-p:
					filename, err := download(part.url, part.filename, part.rangeL, part.rangeR)
					if err != nil {
						results <- result{err: err}
						return
					}
					results <- result{index: part.index, filename: filename}
				default:
					return
				}
			}
		}()
	}
	wg.Wait()

	var files = make([]string, len(parts))
	for {
		select {
		case v := <-results:
			if v.err != nil {
				return nil, v.err
			}
			files[v.index] = v.filename
		default:
			return files, nil
		}
	}
}

func joinFile(files []string, filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, file := range files {
		v, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0600)
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, v); err != nil {
			v.Close()
			return err
		}

		v.Close()
		os.Remove(file)
	}

	return nil
}

func lastFilename(filename string) string {
	i := strings.LastIndex(filename, "/")
	if i == -1 {
		return filename
	}
	return filename[i:]
}
