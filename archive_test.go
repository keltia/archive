package archive

import (
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
	fDebug = false
}

func TestNewArchive_Empty(t *testing.T) {
	a, err := New("")
	require.Error(t, err)
	assert.Empty(t, a)
	assert.IsType(t, (*Plain)(nil), a)
}

func TestNewArchive_Plain(t *testing.T) {
	a, err := New("foo.txt")
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
	a, err := New("foo.zip")
	require.Error(t, err)
	assert.Empty(t, a)
	assert.IsType(t, (*Zip)(nil), a)

}

func TestNewArchive_Gzip(t *testing.T) {
	a, err := New("foo.gz")
	require.NoError(t, err)
	assert.NotEmpty(t, a)
	assert.IsType(t, (*Gzip)(nil), a)
}

func TestNew_Gpg(t *testing.T) {
	a, err := New("foo.asc")
	require.NoError(t, err)
	assert.NotEmpty(t, a)
	assert.IsType(t, (*Gpg)(nil), a)
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
	require.NoError(t, err)
	require.NotNil(t, a)

	txt, err := a.Extract(".txt")
	assert.Error(t, err)
	assert.Empty(t, txt)
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

func TestGpg_Close(t *testing.T) {
	fn := "testdata/notempty.asc"

	a := &Gpg{fn: fn, unc: "notempty.txt", gpg: NullGPG{}}
	require.NoError(t, a.Close())
}
