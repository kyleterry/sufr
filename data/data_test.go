package data

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func SetupTest() string {
	tmpfile, err := ioutil.TempFile("", "sufr-test")
	if err != nil {
		panic(err)
	}

	name := tmpfile.Name()

	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return name
}

func TeardownTest(tmpDB string) {
	err := os.Remove(tmpDB)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	var err error

	tmpDB := SetupTest()

	db, err = New(tmpDB)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	db.Close()
	TeardownTest(tmpDB)

	os.Exit(code)
}

func TestRunWithBucketForType(t *testing.T) {
	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		err = RunWithBucketForType(tx, Settings{}, func(bucket *bolt.Bucket) error {
			assert.NotNil(t, bucket)
			return nil
		})

		assert.NoError(t, err)

		err = RunWithBucketForType(tx, User{}, func(bucket *bolt.Bucket) error {
			assert.NotNil(t, bucket)
			return nil
		})

		assert.NoError(t, err)

		err = RunWithBucketForType(tx, PinnedTag{}, func(bucket *bolt.Bucket) error {
			assert.NotNil(t, bucket)
			return nil
		})

		assert.NoError(t, err)

		err = RunWithBucketForType(tx, PinnedTags{}, func(bucket *bolt.Bucket) error {
			assert.NotNil(t, bucket)
			return nil
		})

		assert.NoError(t, err)

		err = RunWithBucketForType(tx, URL{}, func(bucket *bolt.Bucket) error {
			assert.NotNil(t, bucket)
			return nil
		})

		assert.NoError(t, err)

		err = RunWithBucketForType(tx, Tag{}, func(bucket *bolt.Bucket) error {
			assert.NotNil(t, bucket)
			return nil
		})

		assert.NoError(t, err)

		return err
	})

	assert.NoError(t, err)
}
