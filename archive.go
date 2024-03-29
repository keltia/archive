package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

// Version number (SemVer)
const (
	myVersion = "0.9.1"
)

var (
	fVerbose = false
	fDebug   = false
)

// ------------------- Interfaces

// Extracter is the main interface we have
type Extracter interface {
	Extract(t string) ([]byte, error)
}

// ExtractCloser is the same with Close()
type ExtractCloser interface {
	Extracter
	Close() error
	Type() int
}

const (
	// ArchivePlain starts the different types
	ArchivePlain = 1 << iota
	// ArchiveGzip is for gzip archives
	ArchiveGzip
	// ArchiveZip is for zip archives
	ArchiveZip
	// ArchiveTar describes the tar ones
	ArchiveTar
	// ArchiveGpg is for openpgp archives
	ArchiveGpg
	// ArchiveZstd is for Zstd archives
	ArchiveZstd
)

// ------------------- Plain

// Plain is for plain text
type Plain struct {
	Name string
	r    io.Reader
}

func NewPlainfile(fn string) (*Plain, error) {
	fh, err := os.Open(fn)
	if err != nil {
		return nil, errors.Wrap(err, "NewPlainfile")
	}
	return &Plain{Name: fn, r: fh}, nil
}

// Extract returns the content of the file
func (a Plain) Extract(t string) ([]byte, error) {
	if a.Name == "-" {
		var b bytes.Buffer

		_, err := io.Copy(&b, a.r)
		if err != nil {
			return nil, errors.Wrap(err, "Extract/Copy")
		}
		return b.Bytes(), nil
	}
	ext := filepath.Ext(a.Name)
	if ext == t || t == "" {
		return ioutil.ReadFile(a.Name)
	}
	return []byte{}, fmt.Errorf("wrong file type")
}

// Close is a no-op
func (a Plain) Close() error {
	return nil
}

// Type returns the archive type obviously.
func (a Plain) Type() int {
	return ArchivePlain
}

// ------------------- Zip

// Zip is for pkzip/infozip files
type Zip struct {
	fn  string
	zfh *zip.ReadCloser
}

// NewZipfile open the zip file
func NewZipfile(fn string) (*Zip, error) {
	zfh, err := zip.OpenReader(fn)
	if err != nil {
		return &Zip{}, errors.Wrap(err, "archive/zip")
	}
	return &Zip{fn: fn, zfh: zfh}, nil
}

// Extract returns the content of the file
func (a Zip) Extract(t string) ([]byte, error) {
	verbose("exploring %s", a.fn)

	ft := strings.ToLower(t)
	for _, fn := range a.zfh.File {
		verbose("looking at %s", fn.Name)

		if path.Ext(fn.Name) == ft {
			file, err := fn.Open()
			if err != nil {
				return []byte{}, errors.Wrapf(err, "no file matching type %s", t)
			}
			return ioutil.ReadAll(file)
		}
	}

	return []byte{}, fmt.Errorf("no file matching type %s", t)
}

// Close does something here
func (a Zip) Close() error {
	return a.zfh.Close()
}

// Type returns the archive type obviously.
func (a Zip) Type() int {
	return ArchiveZip
}

// ------------------- Tar

// Tar is a tar archive :)
type Tar struct {
	fn  string
	tfh *tar.Reader
}

func NewTarfile(fn string) (*Tar, error) {
	var fh io.Reader

	if fn == "-" {
		tfh := tar.NewReader(os.Stdin)
		return &Tar{fn: fn, tfh: tfh}, nil
	}

	fh, err := os.Open(fn)
	if err != nil {
		return &Tar{}, errors.Wrap(err, "NewTarfile")
	}

	return &Tar{fn: fn, tfh: tar.NewReader(fh)}, nil
}

func (a Tar) Extract(t string) ([]byte, error) {
	for {
		hdr, err := a.tfh.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return []byte{}, errors.Wrap(err, "read")
		}

		debug("found %s", hdr.Name)

		var buf bytes.Buffer

		if strings.HasSuffix(hdr.Name, t) {
			n, err := io.Copy(&buf, a.tfh)
			if err != nil {
				return []byte{}, errors.Wrap(err, "copy")
			}
			debug("read %d bytes", n)
			return buf.Bytes(), nil
		}
	}
	return nil, errors.New("not found")
}

// Close does something here
func (a Tar) Close() error {
	return nil
}

// Type returns the archive type obviously.
func (a *Tar) Type() int {
	return ArchiveTar
}

// ------------------- Gzip

// Gzip is a gzip-compressed file
type Gzip struct {
	fn  string
	unc string
	gfh io.Reader
}

// NewGzipfile stores the uncompressed file name
func NewGzipfile(fn string) (*Gzip, error) {
	base := filepath.Base(fn)
	pc := strings.Split(base, ".")
	unc := strings.Join(pc[0:len(pc)-1], ".")

	gfh, err := os.Open(fn)
	if err != nil {
		return &Gzip{}, errors.Wrap(err, "NewGzipFile")
	}
	return &Gzip{fn: fn, unc: unc, gfh: gfh}, nil
}

// Extract returns the content of the file
func (a Gzip) Extract(t string) ([]byte, error) {
	zfh, err := gzip.NewReader(a.gfh)
	if err != nil {
		return []byte{}, errors.Wrap(err, "gunzip")
	}
	content, err := ioutil.ReadAll(zfh)
	defer zfh.Close()

	return content, err
}

// Close is a no-op
func (a Gzip) Close() error {
	return nil
}

// Type returns the archive type obviously.
func (a Gzip) Type() int {
	return ArchiveGzip
}

// ------------------- Zstd

// Zstd is a gzip-compressed file
type Zstd struct {
	fn  string
	unc string
	gfh io.Reader
}

// NewZstdfile stores the uncompressed file name
func NewZstdfile(fn string) (*Zstd, error) {
	base := filepath.Base(fn)
	pc := strings.Split(base, ".")
	unc := strings.Join(pc[0:len(pc)-1], ".")

	gfh, err := os.Open(fn)
	if err != nil {
		return &Zstd{}, errors.Wrap(err, "NewZstdFile")
	}
	return &Zstd{fn: fn, unc: unc, gfh: gfh}, nil
}

// Extract returns the content of the file
func (a Zstd) Extract(t string) ([]byte, error) {
	zfh, err := zstd.NewReader(a.gfh)
	if err != nil {
		return []byte{}, errors.Wrap(err, "zstd uncompress")
	}
	content, err := ioutil.ReadAll(zfh)
	defer zfh.Close()

	return content, err
}

// Close is a no-op
func (a Zstd) Close() error {
	return nil
}

// Type returns the archive type obviously.
func (a Zstd) Type() int {
	return ArchiveZstd
}

// ------------------- New/NewFromReader

// New is the main creator
func New(fn string) (ExtractCloser, error) {
	if fn == "" {
		return &Plain{}, fmt.Errorf("null string")
	}
	_, err := os.Stat(fn)
	if err != nil {
		return nil, errors.Wrap(err, "unknown file")
	}
	ext := filepath.Ext(fn)
	switch ext {
	case ".zip":
		return NewZipfile(fn)
	case ".gz":
		return NewGzipfile(fn)
	case ".zst":
		return NewZstdfile(fn)
	case ".asc":
		fallthrough
	case ".gpg":
		return NewGpgfile(fn)
	case ".tar":
		return NewTarfile(fn)
	}
	return NewPlainfile(fn)
}

// NewFromReader uses an io.Reader instead of a file
func NewFromReader(r io.Reader, t int) (ExtractCloser, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}
	fn := "-"
	switch t {
	case ArchivePlain:
		return &Plain{Name: fn, r: r}, nil
	case ArchiveGzip:
		return &Gzip{fn: fn, unc: fn, gfh: r}, nil
	case ArchiveZstd:
		return &Zstd{fn: fn, unc: fn, gfh: r}, nil
	case ArchiveZip:
		return nil, fmt.Errorf("not supported")
	case ArchiveGpg:
		return NewGpgfile(fn)
	case ArchiveTar:
		return NewTarfile(fn)
	}
	return &Plain{Name: fn, r: r}, fmt.Errorf("unknown type")
}

// Ext2Type converts from string to archive type (int)
func Ext2Type(typ string) int {
	switch typ {
	case ".zip":
		return ArchiveZip
	case ".gz":
		return ArchiveGzip
	case ".zst":
		return ArchiveZstd
	case ".asc":
		fallthrough
	case ".gpg":
		return ArchiveGpg
	case ".tar":
		return ArchiveTar
	default:
		return ArchivePlain
	}
}

// ------------------- Misc.

// SetVerbose sets the mode
func SetVerbose() {
	fVerbose = true
}

// SetDebug sets the mode too
func SetDebug() {
	fDebug = true
	fVerbose = true
}

// Reset is for the two flags
func Reset() {
	fDebug = false
	fVerbose = false
}

// Version reports it
func Version() string {
	return myVersion
}
