package config_test

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/chanzuckerberg/reaper/pkg/config"
)

func TestFromFileNoFile(t *testing.T) {
	a := assert.New(t)
	fs := afero.NewMemMapFs()
	_, err := config.FromFile(fs, "foo.yml")
	a.Error(err)
}

func TestFromFileInvalidYaml(t *testing.T) {
	a := assert.New(t)
	fs := afero.NewMemMapFs()
	writeFile(fs, "config.yml", "asdf")

	c, err := config.FromFile(fs, "config.yml")
	a.Nil(c)
	a.Error(err)
}

// lifted from fogg, we need to refactor to go-misc
func writeFile(fs afero.Fs, path string, contents string) error {
	f, e := fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if e != nil {
		return e
	}
	_, e = f.WriteString(contents)
	return e
}
