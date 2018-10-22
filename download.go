package download

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var mb int64 = 1024 * 1024

func Download(url string, filename string, goroutine int) error {
	d := &download{url: url, filename: filename, goroutine: goroutine}

	if err := d.fillValue(); err != nil {
		return err
	}

	if err := d.fillTasks(); err != nil {
		return err
	}

	if err := d.partDownload(); err != nil {
		return err
	}

	if err := d.joinFile(); err != nil {
		return err
	}

	return nil
}

type download struct {
	url       string
	filename  string
	goroutine int

	tempDir       string
	contentLength int64
	lastName      string
	partTasks     []part
	partFiles     []string
}

func (d *download) fillValue() (err error) {
	// last name: xxx.mp4
	if i := strings.LastIndex(d.filename, "/"); i == -1 {
		d.lastName = d.filename
	} else {
		d.lastName = strings.TrimPrefix(d.filename[i:], "/")
	}

	// temp dir
	s := strings.TrimSuffix(d.filename, filepath.Ext(d.filename))
	s = strings.TrimPrefix(s, "/")
	s = strings.Replace(s, "/", "-", -1)
	d.tempDir, err = ioutil.TempDir("", s+"-")
	if err != nil {
		return err
	}

	return nil
}

func (d *download) fillTasks() error {
	// content length
	resp, err := http.Get(d.url)
	if err != nil {
		return err
	}
	d.contentLength = resp.ContentLength

	// part task
	var parts []part
	var i int64 = 0
	for ; (i+1)*mb-1 < d.contentLength; i++ {
		parts = append(parts, part{index: int(i), rangeL: mb * i, rangeR: mb*(i+1) - 1, d: d})
	}
	d.partTasks = parts

	return nil
}

func (d *download) partDownload() error {
	taskLength := len(d.partTasks)
	results := make(chan result, taskLength)
	tasks := make(chan part, taskLength)
	for _, v := range d.partTasks {
		tasks <- v
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < d.goroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case part := <-tasks:
					if err := part.download(d.lastName); err != nil {
						results <- result{err: err}
						return
					}
					results <- result{index: part.index, filename: part.partFilename()}
				default:
					return
				}
			}
		}()
	}
	wg.Wait()

	var files = make([]string, taskLength)
	for {
		select {
		case v := <-results:
			if v.err != nil {
				return v.err
			}
			files[v.index] = v.filename
		default:
			d.partFiles = files
			return nil
		}
	}
	d.partFiles = files
	return nil
}

func (d *download) joinFile() error {
	f, err := os.OpenFile(d.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, file := range d.partFiles {
		v, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0600)
		if err != nil {
			os.Remove(d.filename)
			return err
		}

		if _, err := io.Copy(f, v); err != nil {
			v.Close()
			os.Remove(d.filename)
			return err
		}

		v.Close()
	}

	return os.RemoveAll(d.tempDir)
}
