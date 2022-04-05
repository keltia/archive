// Fake Gpg package to enable compilation on Windows.
//
// XXX there's got to be a better way

//go:build windows || !unix
// +build windows !unix

package archive

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// ------------------- GPG

// Decrypter is the gpgme interface
type Decrypter interface {
	Decrypt(r io.Reader) ([]byte, error)
}

// Gpgme is for real gpgme stuff
type Gpgme struct{}

// Decrypt does the obvious
func (Gpgme) Decrypt(r io.Reader) ([]byte, error) {
	return ioutil.ReadAll(r)
}

// NullGPG is for testing
type NullGPG struct{}

// Decrypt does the obvious
func (NullGPG) Decrypt(r io.Reader) ([]byte, error) {
	return ioutil.ReadAll(r)
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

	verbose("Decrypting %s", a.fn)

	// Do the decryption thing
	plain, err := a.gpg.Decrypt(fh)
	if err != nil {
		return []byte{}, errors.Wrap(err, "extract/decrypt")
	}

	return plain, err
}

// Close is part of the Closer interface
func (a Gpg) Close() error {
	return nil
}

// Type returns the archive type obviously.
func (a *Gpg) Type() int {
	return ArchiveGpg
}
