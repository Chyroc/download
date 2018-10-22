package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// url, /download/test.mp4

// /temp
// /temp/test
// /temp/test/test.mp4-1.cache
// /temp/test/test.mp4-1
// /temp/test/test.mp4-2.cache
// /temp/test/test.mp4-2

// /download/test.mp4

type part struct {
	index  int
	rangeL int64
	rangeR int64

	d *download
}

type result struct {
	index    int
	filename string
	err      error
}

func (p part) getCache() (bool, error) {
	if _, err := os.Stat(p.partFilename()); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if _, err := os.Stat(p.cacheFilename()); err != nil {
		if os.IsNotExist(err) {
			os.Remove(p.partFilename())
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (p part) saveCache() error {
	f, err := os.OpenFile(p.cacheFilename(), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("exist"); err != nil {
		return err
	}

	return nil
}

func (p part) download(last string) error {
	fmt.Println(p)
	// cache
	if exist, err := p.getCache(); err != nil {
		return err
	} else if exist {
		return nil
	}

	// get file
	req, err := http.NewRequest(http.MethodGet, p.d.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", p.rangeL, p.rangeR))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// save
	f, err := os.OpenFile(p.partFilename(), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(p.partFilename())
		os.Remove(p.cacheFilename())
		return err
	}

	f.Close()
	return p.saveCache()
}

func (p part) cacheFilename() string {
	return p.partFilename() + ".cache"
}

func (p part) partFilename() string {
	return p.d.tempDir + "/" + p.d.lastName + "-" + strconv.Itoa(p.index)
}
