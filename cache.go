package download

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func genCacheFilename(s string, index int) string {
	return filepath.Join(os.TempDir(), s+"-"+strconv.Itoa(index)+".cache")
}

func checkCache(cacheFile string) (string, error) {
	bs, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(bs)), nil
}
