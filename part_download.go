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

type part struct {
	index    int
	url      string
	filename string
	rangeL   int64
	rangeR   int64

	cacheFile string
}

type result struct {
	index    int
	filename string
	err      error
}

func (p part) checkCache() (string, error) {
	if p.cacheFile == "" {
		p.cacheFile = genCacheFilename(p.filename, p.index)
	}

	bs, err := ioutil.ReadFile(p.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(bs)), nil
}

func (p part) saveCache(filename string) error {
	if p.cacheFile == "" {
		p.cacheFile = genCacheFilename(p.filename, p.index)
	}

	f, err := os.OpenFile(p.cacheFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(filename); err != nil {
		return err
	}

	return nil
}

func (p part) download() (string, error) {
	f, err := p.checkCache()
	if err != nil {
		return "", err
	} else if f != "" {
		return f, nil
	}

	filename, err := ioutil.TempFile("", p.filename+"-*")
	if err != nil {
		return "", err
	}
	defer filename.Close()

	req, err := http.NewRequest(http.MethodGet, p.url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", p.rangeL, p.rangeR))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(filename, resp.Body); err != nil {
		return "", err
	}

	return filename.Name(), p.saveCache(filename.Name())
}

func partDownload(parts []part, goroutine int) ([]string, error) {
	results := make(chan result, len(parts))
	tasks := make(chan part, len(parts))
	for _, v := range parts {
		tasks <- v
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < goroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case part := <-tasks:
					filename, err := part.download()
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
