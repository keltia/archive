package archive

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/keltia/sandbox"
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

// Gpg

type NullGPGError struct{}

var ErrFakeGPGError = fmt.Errorf("fake error")

func (NullGPGError) Decrypt(r io.Reader) (*gpgme.Data, error) {
	return &gpgme.Data{}, ErrFakeGPGError
}

func TestNullGPG_Decrypt(t *testing.T) {
	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	gpg := NullGPG{}

	file := "testdata/CIMBL-0666-CERTS.zip.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := gpg.Decrypt(fh)
	assert.NoError(t, err)
	assert.Empty(t, plain)
}

func TestNullGPGError_Decrypt(t *testing.T) {
	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	gpg := NullGPGError{}

	file := "testdata/CIMBL-0666-CERTS.zip.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := gpg.Decrypt(fh)
	assert.Error(t, err)
	assert.Equal(t, ErrFakeGPGError, err)
	assert.Empty(t, plain)
}

func TestGpg_Decrypt(t *testing.T) {
	fVerbose = true

	snd, err := sandbox.New("test")
	require.NoError(t, err)
	defer snd.Cleanup()

	gpg := Gpgme{}

	file := "testdata/CIMBL-0666-CERTS.zip.asc"
	fh, err := os.Open(file)
	require.NoError(t, err)
	require.NotNil(t, fh)

	plain, err := gpg.Decrypt(fh)
	assert.Error(t, err)
	assert.NotEmpty(t, plain)
}
