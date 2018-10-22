package download

import (
	"io"
	"os"
	"strings"
)

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
