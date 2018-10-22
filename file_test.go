package download

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestFilename(t *testing.T) {
	as := assert.New(t)

	{
		d := &download{url: "url", filename: "/tmp/1.mp4", goroutine: 10}
		as.Nil(d.fillValue())
		defer os.RemoveAll(d.tempDir)

		as.Equal("/tmp/1.mp4", d.filename)
		as.Equal("1.mp4", d.lastName)
		t.Log(d.tempDir)
		as.True(strings.HasPrefix(d.tempDir, os.TempDir()+"tmp-1-"))
	}

	{
		d := &download{url: "url", filename: "/tmp/2/1.mp4", goroutine: 10}
		as.Nil(d.fillValue())
		defer os.RemoveAll(d.tempDir)

		as.Equal("/tmp/2/1.mp4", d.filename)
		as.Equal("1.mp4", d.lastName)
		t.Log(d.tempDir)
		as.True(strings.HasPrefix(d.tempDir, os.TempDir()+"tmp-2-1-"))
	}

	{
		d := &download{url: "url", filename: "/tmp/2/1.mp4", goroutine: 10}
		as.Nil(d.fillValue())
		p := &part{1, 0, 1, d}

		t.Log(p.cacheFilename())
		as.True(strings.HasPrefix(p.cacheFilename(), os.TempDir()+"tmp-2-1-"))
		as.True(strings.HasSuffix(p.cacheFilename(), "1.mp4-1.cache"))

		t.Log(p.partFilename())
		as.True(strings.HasPrefix(p.partFilename(), os.TempDir()+"tmp-2-1-"))
		as.True(strings.HasSuffix(p.partFilename(), "1.mp4-1"))
	}

	{
		f, err := ensureDirExist("/tmp/3/3.mo4")
		as.Nil(err)
		as.Equal("/tmp/3", f)
		os.Remove("/tmp/3")
	}
}
