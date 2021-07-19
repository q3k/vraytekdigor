package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	mpb "github.com/q3k/vraytekdigor/tarmanip/proto/manipulate"
	"google.golang.org/protobuf/encoding/prototext"
)

var (
	flagScript string
	flagIn     string
	flagOut    string
)

// state of the tarball.
type state struct {
	// map of all files as read from the tarball file and further modified.
	files map[string]*tarEntry
	// file name ordering. Initially read from the source tarball. It's kept
	// original as much as possible, so that even if a file is Removed, Created
	// and then Written, the order of the file within the archive will be as
	// close to the original order as possible.
	names []string
}

// tarEntry is a combined tar entry header and file data.
type tarEntry struct {
	h *tar.Header
	b []byte
}

// readTar reads a source taball into a new state.
func readTar(path string) (*state, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open tarball: %w", err)
	}
	defer in.Close()

	files := make(map[string]*tarEntry)
	var fileNames []string

	tr := tar.NewReader(in)
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read tarball header: %w", err)
		}
		b, err := ioutil.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("could not read tarball file: %w", err)
		}
		te := tarEntry{
			h: h,
			b: b,
		}
		if _, ok := files[h.Name]; ok {
			return nil, fmt.Errorf("duplicate file %q in input", h.Name)
		}
		files[h.Name] = &te
		fileNames = append(fileNames, h.Name)
	}

	return &state{
		files: files,
		names: fileNames,
	}, nil
}

// writeTar emits the (possibly modified) tarball state into a new tarball file.
func (s *state) writeTar(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	for _, p := range s.names {
		fi, ok := s.files[p]
		if !ok {
			continue
		}
		if err := tw.WriteHeader(fi.h); err != nil {
			return fmt.Errorf("failed to write tar file header: %w", err)
		}
		if _, err := tw.Write(fi.b); err != nil {
			return fmt.Errorf("failed to write tar file data: %w", err)
		}
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar file: %w", err)
	}
	return nil
}

// apply a given Script Change object onto the state.
func (s *state) apply(scr *mpb.Script) error {
	for i, change := range scr.Change {
		switch x := change.Kind.(type) {
		case *mpb.Change_Remove:
			if err := s.remove(x.Remove); err != nil {
				return fmt.Errorf("change %d: %w", i, err)
			}
		case *mpb.Change_Write:
			if err := s.write(x.Write); err != nil {
				return fmt.Errorf("change %d: %w", i, err)
			}
		case *mpb.Change_Binreplace:
			if err := s.binreplace(x.Binreplace); err != nil {
				return fmt.Errorf("change %d: %w", i, err)
			}
		case *mpb.Change_Create:
			if err := s.create(x.Create); err != nil {
				return fmt.Errorf("change %d: %w", i, err)
			}
		case *mpb.Change_Jsonpatch:
			if err := s.jsonpatch(x.Jsonpatch); err != nil {
				return fmt.Errorf("change %d: %w", i, err)
			}
		default:
			return fmt.Errorf("unimplemented %v", change.Kind)
		}
	}
	return nil
}

// subfiles returns a list of filenames under the given path.
// TODO(q3k): make this faster?
func (s *state) subfiles(path string) []string {
	var res []string
	for k, _ := range s.files {
		if strings.HasPrefix(k, path) {
			res = append(res, k)
		}
	}
	return res
}

func main() {
	flag.StringVar(&flagScript, "script", "", "Path to script prototext")
	flag.StringVar(&flagIn, "in", "", "Path to input tarball")
	flag.StringVar(&flagOut, "out", "", "Path to input tarball")
	flag.Parse()

	if flagScript == "" {
		log.Fatalf("-script must be set")
	}
	if flagIn == "" {
		log.Fatalf("-in must be set")
	}
	if flagOut == "" {
		log.Fatalf("-out must be set")
	}

	data, err := ioutil.ReadFile(flagScript)
	if err != nil {
		log.Fatalf("Could not read script file: %v", err)
	}
	var s mpb.Script
	if err := prototext.Unmarshal(data, &s); err != nil {
		log.Fatalf("Could not parse script: %v", err)
	}

	st, err := readTar(flagIn)
	if err != nil {
		log.Fatalf("Could not load input: %v", err)
	}

	if err := st.apply(&s); err != nil {
		log.Fatalf("Could not apply script: %v", err)
	}

	if err := st.writeTar(flagOut); err != nil {
		log.Fatalf("Could not write output: %v", err)
	}
}
