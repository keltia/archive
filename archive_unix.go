//go:build unix || !windows
// +build unix !windows

package archive

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/proglottis/gpgme"
)

// ------------------- GPG

// Decrypter is the gpgme interface
type Decrypter interface {
	Decrypt(r io.Reader) (*gpgme.Data, error)
}

// Gpgme is for real gpgme stuff
type Gpgme struct{}

// Decrypt does the obvious
func (Gpgme) Decrypt(r io.Reader) (*gpgme.Data, error) {
	return gpgme.Decrypt(r)
}

// NullGPG is for testing
type NullGPG struct{}

// Decrypt does the obvious
func (NullGPG) Decrypt(r io.Reader) (*gpgme.Data, error) {
	b, _ := ioutil.ReadAll(r)
	return gpgme.NewDataBytes(b)
}

// Gpg is how we use/mock decryption stuff
type Gpg struct {
	fn  string
	unc string
	gpg Decrypter
}

// NewGpgfile initializes the struct and check filename
func NewGpgfile(fn string) (*Gpg, error) {
	// Strip .gpg or .asc from filename
	base := filepath.Base(fn)
	pc := strings.Split(base, ".")
	unc := strings.Join(pc[0:len(pc)-1], ".")

	return &Gpg{fn: fn, unc: unc, gpg: Gpgme{}}, nil
}

// Extract binds it to the Archiver interface
func (a Gpg) Extract(t string) ([]byte, error) {
	// Carefully open the box
	fh, err := os.Open(a.fn)
	if err != nil {
		return []byte{}, errors.Wrap(err, "extract/open")
	}
	defer fh.Close()

	var buf bytes.Buffer

	// Do the decryption thing
	plain, err := a.gpg.Decrypt(fh)
	if err != nil {
		return []byte{}, errors.Wrap(err, "extract/decrypt")
	}
	defer plain.Close()

	// Save "plain" text

	verbose("Decrypting %s", a.fn)

	_, err = io.Copy(&buf, plain)
	if err != nil {
		return []byte{}, errors.Wrap(err, "extract/copy")
	}

	return buf.Bytes(), err
}

// Close is part of the Closer interface
func (a Gpg) Close() error {
	return nil
}

// Type returns the archive type obviously.
func (a *Gpg) Type() int {
	return ArchiveGpg
}
