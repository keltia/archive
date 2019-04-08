package archive

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/proglottis/gpgme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Misc.

func TestVersion(t *testing.T) {
	require.Equal(t, myVersion, Version())
}

func TestSetVerbose(t *testing.T) {
	assert.False(t, fVerbose)
	SetVerbose()
	assert.True(t, fVerbose)
	fVerbose = false
}

func TestSetDebug(t *testing.T) {
	assert.False(t, fDebug)
	SetDebug()
	assert.True(t, fDebug)
	assert.True(t, fVerbose)
	fDebug = false
	fVerbose = false
}

func TestReset(t *testing.T) {
	fVerbose = true
	Reset()
	require.False(t, fVerbose)
}

// NewArchive

func TestNewArchive_Empty(t *testing.T) {
	a, err := New("")
	require.Error(t, err)
	assert.Empty(t, a)
	assert.IsType(t, (*Plain)(nil), a)
}

func TestNewArchive_None(t *testing.T) {
	_, err := New("foo.txt")
	require.Error(t, err)
}

func TestNewArchive_Plain(t *testing.T) {
	a, err := New("testdata/empty.txt")
	require.NoError(t, err)
	assert.NotEmpty(t, a)
	assert.IsType(t, (*Plain)(nil), a)
}

func TestNewArchive_Zip(t *testing.T) {
	a, err := New("testdata/notempty.zip")
	require.NoError(t, err)
	assert.NotEmpty(t, a)
	assert.IsType(t, (*Zip)(nil), a)

}

func TestNewArchive_ZipNone(t *testing.T) {
	a, err := New("testdata/foo.zip")
	require.Error(t, err)
	assert.Empty(t, a)
}

func TestNewArchive_Gzip(t *testing.T) {
	a, err := New("testdata/notempty.txt.gz")
	require.NoError(t, err)
	assert.NotEmpty(t, a)
	assert.IsType(t, (*Gzip)(nil), a)
}

func TestNewArchive_Gpg(t *testing.T) {
	a, err := New("testdata/notempty.asc")
	require.NoError(t, err)
	assert.NotEmpty(t, a)
	assert.IsType(t, (*Gpg)(nil), a)
}

func TestNewArchive_Tar(t *testing.T) {
	a, err := New("testdata/notempty.tar")
	require.NoError(t, err)
	assert.NotEmpty(t, a)
	assert.IsType(t, (*Tar)(nil), a)
}

// Plain

func TestPlain_Extract(t *testing.T) {
	fn := "testdata/notempty.txt"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	txt, err := a.Extract("")
	assert.NoError(t, err)
	assert.Equal(t, "this is a file\n", string(txt))
}

func TestPlain_Extract2(t *testing.T) {
	fn := "testdata/notempty.txt"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	txt, err := a.Extract(".doc")
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestPlain_Close(t *testing.T) {
	fn := "testdata/notempty.txt"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	require.NoError(t, a.Close())
}

// Zip

func TestNewZipfile_Garbage(t *testing.T) {
	fn := "testdata/garbage.zip"
	a, err := NewZipfile(fn)
	require.Error(t, err)
	assert.Empty(t, a)
}

func TestZip_Extract(t *testing.T) {
	fn := "testdata/notempty.zip"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()

	rh, err := ioutil.ReadFile("testdata/notempty.txt")
	require.NoError(t, err)
	require.NotEmpty(t, rh)

	txt, err := a.Extract(".txt")
	assert.NoError(t, err)
	assert.Equal(t, string(rh), string(txt))
}

func TestZip_Extract2(t *testing.T) {
	fn := "testdata/notempty.zip"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	txt, err := a.Extract(".xml")
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestZip_Close(t *testing.T) {
	fn := "testdata/notempty.zip"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	require.NoError(t, a.Close())
}

// Gzip

func TestGzip_Extract(t *testing.T) {
	fn := "testdata/notempty.txt.gz"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()

	rh, err := ioutil.ReadFile("testdata/notempty.txt")
	require.NoError(t, err)
	require.NotEmpty(t, rh)

	txt, err := a.Extract(".txt")
	t.Logf("err=%v", err)
	assert.NoError(t, err)
	assert.Equal(t, string(rh), string(txt))
}

func TestGzip_Extract2(t *testing.T) {
	fn := "testdata/notempty.txt.gz"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	txt, err := a.Extract(".xml")
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestGzip_Extract3(t *testing.T) {
	fn := "/nonexistent"
	a, err := New(fn)
	require.Error(t, err)
	require.Empty(t, a)
}

func TestGzip_Close(t *testing.T) {
	fn := "testdata/notempty.txt.gz"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	require.NoError(t, a.Close())
}

// gpg

type NullGPGError struct{}

var ErrFakeGPGError = fmt.Errorf("fake error")

func (NullGPGError) Decrypt(r io.Reader) (*gpgme.Data, error) {
	b, _ := ioutil.ReadAll(r)
	gb, _ := gpgme.NewDataBytes(b)
	return gb, ErrFakeGPGError
}

func TestNullGPG_Decrypt(t *testing.T) {
	fVerbose = true

	gpg := NullGPG{}

	file := "testdata/notempty.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := gpg.Decrypt(fh)
	assert.NoError(t, err)
	assert.NotEmpty(t, plain)

	var buf strings.Builder

	_, err = io.Copy(&buf, plain)
	assert.NoError(t, err)
	assert.Equal(t, len("this is a file\n"), len(buf.String()))
	assert.Equal(t, "this is a file\n", buf.String())
}

func TestNullGPGError_Decrypt(t *testing.T) {
	fVerbose = true

	gpg := NullGPGError{}

	file := "testdata/notempty.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := gpg.Decrypt(fh)
	assert.Error(t, err)
	assert.Equal(t, ErrFakeGPGError, err)

	var buf strings.Builder

	_, err = io.Copy(&buf, plain)
	assert.NoError(t, err)
	assert.Equal(t, len("this is a file\n"), len(buf.String()))
	assert.Equal(t, "this is a file\n", buf.String())
}

func TestGpg_Decrypt(t *testing.T) {
	fVerbose = true

	gpg := Gpgme{}

	file := "testdata/notempty.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := gpg.Decrypt(fh)
	assert.Error(t, err)
	assert.NotEmpty(t, plain)
}

func TestGpg_Extract(t *testing.T) {
	fn := "testdata/notempty.asc"

	a := &Gpg{fn: fn, unc: "notempty.txt", gpg: NullGPG{}}
	defer a.Close()

	rh, err := ioutil.ReadFile("testdata/notempty.txt")
	require.NoError(t, err)
	require.NotEmpty(t, rh)

	txt, err := a.Extract(".txt")
	assert.NoError(t, err)
	assert.Equal(t, string(rh), string(txt))
}

func TestGpg_Extract2(t *testing.T) {
	fn := "testdata/notempty.asc"

	a := &Gpg{fn: fn, unc: "notempty.txt", gpg: Gpgme{}}
	defer a.Close()

	_, err := a.Extract(".txt")
	assert.Error(t, err)
}

func TestGpg_Extract3(t *testing.T) {
	fn := "testdata/notempty.nowhere"

	a := &Gpg{fn: fn, unc: "notempty.txt", gpg: Gpgme{}}
	defer a.Close()

	_, err := a.Extract(".txt")
	assert.Error(t, err)
}

func TestGpg_Extract4(t *testing.T) {
	fn := "testdata/notempty.zip.asc"

	a := &Gpg{fn: fn, unc: "notempty.zip", gpg: NullGPG{}}
	defer a.Close()

	rh, err := ioutil.ReadFile("testdata/notempty.zip")
	require.NoError(t, err)
	require.NotEmpty(t, rh)

	zip, err := a.Extract(".zip")
	assert.NoError(t, err)
	assert.Equal(t, string(rh), string(zip))
}

func TestGpg_Extract4_Debug(t *testing.T) {
	fn := "testdata/notempty.zip.asc"
	fDebug = true

	a := &Gpg{fn: fn, unc: "notempty.zip", gpg: NullGPG{}}
	defer a.Close()

	rh, err := ioutil.ReadFile("testdata/notempty.zip")
	require.NoError(t, err)
	require.NotEmpty(t, rh)

	zip, err := a.Extract(".zip")
	assert.NoError(t, err)
	assert.Equal(t, string(rh), string(zip))
	fDebug = false
}

func TestGpg_Close(t *testing.T) {
	fn := "testdata/notempty.asc"

	a := &Gpg{fn: fn, unc: "notempty.txt", gpg: NullGPG{}}
	require.NoError(t, a.Close())
}

// Tar

func TestNewTarfile(t *testing.T) {
	fn := "/nonexistent"
	a, err := NewTarfile(fn)
	require.Error(t, err)
	assert.Empty(t, a)
}

func TestTar_Extract(t *testing.T) {
	fn := "testdata/notempty.tar"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()

	rh, err := ioutil.ReadFile("testdata/notempty.txt")
	require.NoError(t, err)
	require.NotEmpty(t, rh)

	txt, err := a.Extract("notempty.txt")
	assert.NoError(t, err)
	assert.Equal(t, string(rh), string(txt))
}

func TestTar_Extract_Debug(t *testing.T) {
	fn := "testdata/notempty.tar"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()

	SetDebug()
	rh, err := ioutil.ReadFile("testdata/notempty.txt")
	require.NoError(t, err)
	require.NotEmpty(t, rh)

	txt, err := a.Extract("notempty.txt")
	assert.NoError(t, err)
	assert.Equal(t, string(rh), string(txt))
	Reset()
}

func TestTar_Extract_Empty(t *testing.T) {
	fn := "testdata/empty.tar"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)
	defer a.Close()

	txt, err := a.Extract(".txt")
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestTar_Extract2(t *testing.T) {
	fn := "testdata/notempty.tar"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	txt, err := a.Extract(".xml")
	assert.Error(t, err)
	assert.Empty(t, txt)
}

func TestTar_Close(t *testing.T) {
	fn := "testdata/notempty.tar"
	a, err := New(fn)
	require.NoError(t, err)
	require.NotNil(t, a)

	require.NoError(t, a.Close())
}

// FromReader

func TestNewFromReader_Nil(t *testing.T) {
	a, err := NewFromReader(nil, ArchivePlain)
	assert.Error(t, err)
	assert.Empty(t, a)
}

func TestNewFromReader_Plain(t *testing.T) {
	cipher, err := ioutil.ReadFile("testdata/notempty.txt")
	assert.NoError(t, err)
	assert.NotEmpty(t, cipher)

	var buf bytes.Buffer

	n, err := buf.Write(cipher)
	assert.NoError(t, err)
	assert.Equal(t, len(cipher), n)

	a, err := NewFromReader(&buf, ArchivePlain)
	assert.NoError(t, err)
	assert.NotEmpty(t, a)
	require.Equal(t, ArchivePlain, a.Type())
}

func TestNewFromReader_Gpg(t *testing.T) {
	cipher, err := ioutil.ReadFile("testdata/notempty.asc")
	assert.NoError(t, err)
	assert.NotEmpty(t, cipher)

	var buf bytes.Buffer

	n, err := buf.Write(cipher)
	assert.NoError(t, err)
	assert.Equal(t, len(cipher), n)

	a, err := NewFromReader(&buf, ArchiveGpg)
	assert.NoError(t, err)
	assert.NotEmpty(t, a)
	require.Equal(t, ArchiveGpg, a.Type())
}

func TestNewFromReader_Zip(t *testing.T) {
	file, err := ioutil.ReadFile("testdata/notempty.zip")
	assert.NoError(t, err)
	assert.NotEmpty(t, file)

	var buf bytes.Buffer

	n, err := buf.Write(file)
	assert.NoError(t, err)
	assert.Equal(t, len(file), n)

	a, err := NewFromReader(&buf, ArchiveZip)
	assert.Error(t, err)
	assert.Empty(t, a)
}

func TestNewFromReader_Gzip(t *testing.T) {
	file, err := ioutil.ReadFile("testdata/notempty.txt.gz")
	assert.NoError(t, err)
	assert.NotEmpty(t, file)

	var buf bytes.Buffer

	n, err := buf.Write(file)
	assert.NoError(t, err)
	assert.Equal(t, len(file), n)

	a, err := NewFromReader(&buf, ArchiveGzip)
	assert.NoError(t, err)
	assert.NotEmpty(t, a)
	require.Equal(t, ArchiveGzip, a.Type())
}

func TestNewFromReader_Tar(t *testing.T) {
	file, err := ioutil.ReadFile("testdata/notempty.tar")
	assert.NoError(t, err)
	assert.NotEmpty(t, file)

	var buf bytes.Buffer

	n, err := buf.Write(file)
	assert.NoError(t, err)
	assert.Equal(t, len(file), n)

	a, err := NewFromReader(&buf, ArchiveTar)
	assert.NoError(t, err)
	assert.NotEmpty(t, a)
	require.Equal(t, ArchiveTar, a.Type())
}

func TestNewFromReader_Invalid(t *testing.T) {
	file, err := ioutil.ReadFile("testdata/notempty.tar")
	assert.NoError(t, err)
	assert.NotEmpty(t, file)

	var buf bytes.Buffer

	n, err := buf.Write(file)
	assert.NoError(t, err)
	assert.Equal(t, len(file), n)

	a, err := NewFromReader(&buf, 666)
	assert.Error(t, err)
	assert.NotEmpty(t, a)
	assert.Equal(t, &Plain{"-"}, a)
}

func TestExt2Type(t *testing.T) {
	td := []struct {
		ins string
		out int
	}{
		{"", ArchivePlain},
		{".zip", ArchiveZip},
		{".gz", ArchiveGzip},
		{".asc", ArchiveGpg},
		{".gpg", ArchiveGpg},
		{".tar", ArchiveTar},
		{".txt", ArchivePlain},
	}

	for _, d := range td {
		assert.Equal(t, d.out, Ext2Type(d.ins))
	}
}
