package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"syreclabs.com/go/faker"
)

func TestCopy(t *testing.T) {
	nameFileSource, nameFileDestination, limit, offset, err := generateTempFile()
	require.NoError(t, err)
	fileSource, err := os.Open(nameFileSource)
	fs, _ := fileSource.Stat()
	sizeFileSource := fs.Size()
	fileSource.Close()
	defer os.Remove(nameFileDestination)
	defer os.Remove(nameFileSource)
	t.Run("Full Copy", func(t *testing.T) {
		err := Copy(nameFileSource, nameFileDestination, 0, 0)
		require.NoError(t, err)
		expectedBuf, err := ioutil.ReadFile(nameFileSource)
		require.NoError(t, err)
		actualBuf, err := ioutil.ReadFile(nameFileDestination)
		require.NoError(t, err)
		require.Equal(t, expectedBuf, actualBuf)
	})
	t.Run("Copy with  limit", func(t *testing.T) {
		err = Copy(nameFileSource, nameFileDestination, 0, limit)
		require.NoError(t, err)
		expectedBuf, err := ioutil.ReadFile(nameFileSource)
		require.NoError(t, err)
		actualBuf, err := ioutil.ReadFile(nameFileDestination)
		require.NoError(t, err)
		require.Equal(t, expectedBuf[:limit], actualBuf)
	})
	t.Run("Copy with  offset", func(t *testing.T) {
		err = Copy(nameFileSource, nameFileDestination, offset, 0)
		require.NoError(t, err)
		expectedBuf, err := ioutil.ReadFile(nameFileSource)
		require.NoError(t, err)
		actualBuf, err := ioutil.ReadFile(nameFileDestination)
		require.NoError(t, err)
		require.Equal(t, expectedBuf[offset:], actualBuf)
	})
	t.Run("Copy with  offset and limit", func(t *testing.T) {
		err = Copy(nameFileSource, nameFileDestination, offset, limit)
		require.NoError(t, err)
		expectedBuf, err := ioutil.ReadFile(nameFileSource)
		require.NoError(t, err)
		actualBuf, err := ioutil.ReadFile(nameFileDestination)
		require.NoError(t, err)
		require.Equal(t, expectedBuf[offset:][:limit], actualBuf)
	})
	t.Run("Copy with  offset and huge limit", func(t *testing.T) {
		err = Copy(nameFileSource, nameFileDestination, offset, sizeFileSource)
		require.NoError(t, err)
		expectedBuf, err := ioutil.ReadFile(nameFileSource)
		require.NoError(t, err)
		actualBuf, err := ioutil.ReadFile(nameFileDestination)
		require.NoError(t, err)
		require.Equal(t, expectedBuf[offset:], actualBuf)
	})
	t.Run("Error offset", func(t *testing.T) {
		err = Copy(nameFileSource, nameFileDestination, sizeFileSource, 0)
		require.Equal(t, err, ErrOffsetExceedsFileSize)
	})
}

func encode(s []string) []byte {
	var b []byte
	//b = writeLen(b, len(s))
	for _, ss := range s {
		//b = writeLen(b, len(ss))
		b = append(b, ss...)
	}
	return b
}

func generateTempFile() (nameFileSource, nameFileDestination string, offset, limit int64, err error) {
	var fileSourse *os.File
	var t []string
	for {
		t = faker.Hacker().Phrases()
		if len(t) > 4 {
			break
		}
	}
	nameFileSource = faker.App().Name() + ".txt"
	nameFileDestination = faker.App().Name() + ".txt"
	offset = int64(len([]byte(t[0])))
	limit = int64(len([]byte(t[1])))
	if fileSourse, err = os.Create(nameFileSource); err != nil {
		return
	}
	defer func() {
		fileSourse.Close()
	}()
	if _, err = fileSourse.Write(encode(t)); err != nil {
		return
	}
	return
}
