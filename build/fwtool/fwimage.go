package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"log"
	"os"

	"github.com/lunixbochs/struc"
)

type header struct {
	// Magic field, expected 2RDH (HDR2?)
	Magic []byte `struc:"[4]byte,big"`
	// Header size, expected 256 bytes.
	HeaderSize uint32 `struc:"uint32,big"`
	// File size without footer.
	FileSize uint32 `struc:"uint32,big"`
	// CRC32 of data between header and footer.
	CRC      uint32 `struc:"uint32,big"`
	Padding0 []byte `struc:"[64]byte"`
	// Size of kernel blob, and thus also offset of SquashFS blob within data.
	KernelSize uint32 `struc:"uint32,big"`
	// Size of SquashFS blob.
	SquashFSSize uint32 `struc:"uint32,big"`
	Padding1     []byte `struc:"[49]byte"`
	// Checked by firmware update process, seems to be always zeroes?
	Unknown0 uint8
	// Checked by firmware update process, seems to be always zeroes?
	Unknown1 uint8
	Padding2 []byte `struct:"[117]byte"`
}

type footer struct {
	// Magic field, expected DrayTekImageMD5\n.
	Magic []byte `struc:"[16]byte"`
	// Hex-encoded MD5-sum of file up to (ie. header.FileSize bytes long),
	// newline-terminated.
	Sum []byte `struc:"[33]byte"`
}

type fwFile struct {
	h      header
	f      footer
	kernel []byte
	squash []byte
}

func parseFWFile(path string) (*fwFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open firmware file for reading: %w", err)
	}
	defer f.Close()

	// Read and verify header.
	h := header{}
	if err := struc.Unpack(f, &h); err != nil {
		return nil, fmt.Errorf("could not parse firmware header: %w", err)
	}

	log.Printf("Firmware: hsize: %x, fsize: %x, crc: %x, ksize: %x, ssize: %x",
		h.HeaderSize, h.FileSize, h.CRC, h.KernelSize, h.SquashFSSize)

	if want, got := []byte("2RDH"), h.Magic; !bytes.Equal(want, got) {
		return nil, fmt.Errorf("header magic invalid, wanted %q got %q", want, got)
	}
	if want, got := uint32(0x100), h.HeaderSize; want != got {
		return nil, fmt.Errorf("header size invalid, wanted %d got %d", want, got)
	}

	if want, got := h.HeaderSize+h.KernelSize+h.SquashFSSize, h.FileSize; want != got {
		return nil, fmt.Errorf("calculated file size invalid, wanted %x, got %x", want, got)
	}

	// Read payloads (kernel and squashfs).
	kernel := make([]byte, h.KernelSize)
	if _, err := f.Read(kernel); err != nil {
		return nil, fmt.Errorf("failed to read kernel: %w", err)
	}
	squash := make([]byte, h.SquashFSSize)
	if _, err := f.Read(squash); err != nil {
		return nil, fmt.Errorf("failed to read squashfs: %w", err)
	}

	// Read and verify footer.
	fr := footer{}
	if err := struc.Unpack(f, &fr); err != nil {
		return nil, fmt.Errorf("could not parse firmware footer: %w", err)
	}

	if want, got := []byte("DrayTekImageMD5\n"), fr.Magic; !bytes.Equal(want, got) {
		return nil, fmt.Errorf("footer magic invalid, wanted %q, got %q", want, got)
	}
	if want, got := byte(0x0a), fr.Sum[32]; want != got {
		return nil, fmt.Errorf("footer trailing byte invalid, wanted newline, got %x", got)
	}

	// Re-read entire file for checksumming.
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to beginning of file: %w", err)
	}
	data := make([]byte, h.FileSize)
	if _, err := f.Read(data); err != nil {
		return nil, fmt.Errorf("failed to read file for checksumming: %w", err)
	}

	// Calculate and check CRC32.
	if want, got := crc32.ChecksumIEEE(data[h.HeaderSize:])^0xffffffff, h.CRC; want != got {
		return nil, fmt.Errorf("header checksum invalid, wanted %x, got %x", want, got)
	}
	// Calculate and check MD5.
	sum := md5.Sum(data)
	if want, got := hex.EncodeToString(sum[:]), string(fr.Sum[:32]); want != got {
		return nil, fmt.Errorf("footer digest invalid, wanted %s, got %s", want, got)
	}

	log.Printf("Firmware: CRC and MD5 okay.")

	return &fwFile{
		h:      h,
		f:      fr,
		kernel: kernel,
		squash: squash,
	}, nil
}

func (f *fwFile) write(path string) error {
	if want, got := 0x0021e1b5, len(f.kernel); got != want {
		return fmt.Errorf("kernel blob wrong size, got %x bytes, wanted %x", got, want)
	}
	if want, got := 0x00ed0000, len(f.squash); got > want {
		return fmt.Errorf("squashfs blob too large, got %x bytes, max %x", got, want)
	}

	f.h.KernelSize = uint32(len(f.kernel))
	f.h.SquashFSSize = uint32(len(f.squash))
	f.h.FileSize = f.h.HeaderSize + f.h.KernelSize + f.h.SquashFSSize
	c := crc32.NewIEEE()
	c.Write(f.kernel)
	c.Write(f.squash)
	f.h.CRC = c.Sum32() ^ 0xffffffff

	log.Printf("Writing: crc: %x, ksize: %x, ssize: %x", f.h.CRC, f.h.KernelSize, f.h.SquashFSSize)

	o, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	s := md5.New()

	if err := struc.Pack(o, f.h); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if err := struc.Pack(s, f.h); err != nil {
		return fmt.Errorf("failed to hash header: %w", err)
	}
	if _, err := o.Write(f.kernel); err != nil {
		return fmt.Errorf("failed to write kernel: %w", err)
	}
	if _, err := s.Write(f.kernel); err != nil {
		return fmt.Errorf("failed to hash kernel: %w", err)
	}
	if _, err := o.Write(f.squash); err != nil {
		return fmt.Errorf("failed to write squashfs: %w", err)
	}
	if _, err := s.Write(f.squash); err != nil {
		return fmt.Errorf("failed to hash squashfs: %w", err)
	}

	sum := s.Sum(nil)

	if _, err := fmt.Fprintf(o, "DrayTekImageMD5\n%s\n", hex.EncodeToString(sum[:])); err != nil {
		return fmt.Errorf("failed to write footer: %w", err)
	}

	return nil
}
