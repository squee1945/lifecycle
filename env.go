package lifecycle

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Env struct {
	Getenv  func(key string) string
	Setenv  func(key, value string) error
	Environ func() []string
	Map     map[string][]string
}

func (p *Env) AddRootDir(baseDir string) error {
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}
	for dir, vars := range p.Map {
		newDir := filepath.Join(absBaseDir, dir)
		if _, err := os.Stat(newDir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		for _, key := range vars {
			value := newDir + prefix(p.Getenv(key), os.PathListSeparator)
			if err := p.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Env) AddEnvDir(envDir string) error {
	return eachEnvFile(envDir, func(k, v string) error {
		parts := strings.SplitN(k, ".", 2)
		name := parts[0]
		var action string
		if len(parts) > 1 {
			action = parts[1]
		}
		switch action {
		case "prepend":
			return p.Setenv(name, v+prefix(p.Getenv(name), delim(envDir, name)...))
		case "append":
			return p.Setenv(name, suffix(p.Getenv(name), delim(envDir, name)...)+v)
		case "override":
			return p.Setenv(name, v)
		case "default":
			if p.Getenv(name) != "" {
				return nil
			}
			return p.Setenv(name, v)
		case "":
			return p.Setenv(name, v+prefix(p.Getenv(name), delim(envDir, name, os.PathListSeparator)...))
		default:
			return nil
		}
	})
}

func (p *Env) List() []string {
	return p.Environ()
}

func prefix(s string, prefix ...byte) string {
	if s == "" {
		return ""
	}
	return string(prefix) + s
}

func suffix(s string, suffix ...byte) string {
	if s == "" {
		return ""
	}
	return s + string(suffix)
}

func delim(dir, name string, def ...byte) []byte {
	value, err := ioutil.ReadFile(filepath.Join(dir, name+".delim"))
	if err != nil {
		return def
	}
	return value
}

func eachEnvFile(dir string, fn func(k, v string) error) error {
	files, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		value, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return err
		}
		if err := fn(f.Name(), string(value)); err != nil {
			return err
		}
	}
	return nil
}
