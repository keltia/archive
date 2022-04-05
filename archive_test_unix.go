//go:build unix || !windows
// +build unix !windows

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
