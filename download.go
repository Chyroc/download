package download

import (
	"net/http"
)

var mb int64 = 1024 * 1024

func Download(url string, filename string, goroutine int) error {
	contentLength, err := getContentLength(url)
	if err != nil {
		return err
	}

	parts := genPartsTask(url, filename, contentLength)

	files, err := partDownload(parts, goroutine)
	if err != nil {
		return err
	}

	return joinFile(files, filename)
}

func getContentLength(url string) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	return resp.ContentLength, nil
}

func genPartsTask(url, filename string, contentLength int64) (parts []part) {
	last := lastFilename(filename)

	var i int64 = 0
	for ; (i+1)*mb-1 < contentLength; i++ {
		parts = append(parts, part{index: int(i), url: url, filename: last, rangeL: mb * i, rangeR: mb*(i+1) - 1})
	}
	return parts
}
