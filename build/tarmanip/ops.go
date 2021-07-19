package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	jsonpatch "github.com/evanphx/json-patch"
	mpb "github.com/q3k/vraytekdigor/tarmanip/proto/manipulate"
)

// ensures that the given path exists and returns its corresponding tarEntry.
func (s *state) mustExist(path string) (*tarEntry, error) {
	if path == "" {
		return nil, fmt.Errorf("not set")
	}
	if path[0] != '/' {
		return nil, fmt.Errorf("must be absolute")
	}
	fi, ok := s.files[path[1:]]
	if !ok {
		return nil, fmt.Errorf("%s not found in file", path)
	}
	return fi, nil
}

func (s *state) remove(r *mpb.Remove) error {
	f, err := s.mustExist(r.Path)
	if err != nil {
		return fmt.Errorf("path %w", err)
	}

	if f.h.Typeflag == tar.TypeReg {
		delete(s.files, r.Path[1:])
		log.Printf("Removed file %s", r.Path)
		return nil
	} else {
		subfiles := s.subfiles(r.Path[1:])
		if len(subfiles) > 0 {
			if !r.Recursive {
				return fmt.Errorf("path %q not empty, mark as recursive?", r.Path)
			}
			for _, sub := range subfiles {
				delete(s.files, sub)
				log.Printf("Removed (recursively) /%s", sub)
			}
		}
		delete(s.files, r.Path[1:])
		log.Printf("Removed directory %s", r.Path)
	}
	return nil
}

func (s *state) write(w *mpb.Write) error {
	fi, err := s.mustExist(w.Path)
	if err != nil {
		return fmt.Errorf("path %w", err)
	}

	if w.Source == "" {
		return fmt.Errorf("source must be set")
	}
	data, err := ioutil.ReadFile(w.Source)
	if err != nil {
		return fmt.Errorf("could not read source: %w", err)
	}
	fi.b = data
	fi.h.Size = int64(len(data))
	return nil
}

func (s *state) binreplace(b *mpb.BinReplace) error {
	fi, err := s.mustExist(b.Path)
	if err != nil {
		return fmt.Errorf("path %w", err)
	}

	if b.From == "" {
		return fmt.Errorf("from must be set")
	}
	if len(b.To) > len(b.From) {
		return fmt.Errorf("to longer then from")
	}

	padByte := []byte(b.Pad)
	switch len(padByte) {
	case 0:
		padByte = []byte("\x00")
	case 1:
	default:
		return fmt.Errorf("pad byte must be one byte long")
	}
	to := []byte(b.To)
	pad := len(b.From) - len(b.To)
	to = append(to, bytes.Repeat(padByte, pad)...)

	replaced := bytes.ReplaceAll(fi.b, []byte(b.From), to)
	if bytes.Equal(replaced, fi.b) {
		return fmt.Errorf("from not found in file data")
	}
	fi.b = replaced
	return nil
}

func (s *state) create(c *mpb.Create) error {
	if c.Path == "" {
		return fmt.Errorf("path must be set")
	}
	if c.Path[0] != '/' {
		return fmt.Errorf("path must be absolute")
	}
	path := c.Path[1:]
	if _, ok := s.files[path]; ok {
		return fmt.Errorf("path %s already exists", c.Path)
	}

	if c.Mode == 0 {
		c.Mode = 0644
	}

	s.files[path] = &tarEntry{
		h: &tar.Header{
			Name: path,
			Mode: int64(c.Mode),
			Size: 0,
		},
	}
	exists := false
	for _, n := range s.names {
		if n == path {
			exists = true
			break
		}
	}
	if !exists {
		s.names = append(s.names, path)
	}
	return nil
}

func (s *state) jsonpatch(p *mpb.JSONPatch) error {
	fi, err := s.mustExist(p.Path)
	if err != nil {
		return fmt.Errorf("path %w", err)
	}

	if p.Source == "" {
		return fmt.Errorf("source must be set")
	}
	data, err := ioutil.ReadFile(p.Source)
	if err != nil {
		return fmt.Errorf("could not read source: %w", err)
	}

	patch, err := jsonpatch.DecodePatch(data)
	if err != nil {
		return fmt.Errorf("could not decide patch: %w", err)
	}
	modified, err := patch.Apply(fi.b)
	if err != nil {
		return fmt.Errorf("could not apply patch: %w", err)
	}
	fi.b = modified
	fi.h.Size = int64(len(fi.b))
	return nil
}
