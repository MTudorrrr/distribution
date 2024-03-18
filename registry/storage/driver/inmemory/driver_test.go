package inmemory

import (
	"testing"

	storagedriver "github.com/MTudorrrr/distribution/registry/storage/driver"
	"github.com/MTudorrrr/distribution/registry/storage/driver/testsuites"
)

func newDriverConstructor() (storagedriver.StorageDriver, error) {
	return New(), nil
}

func TestInMemoryDriverSuite(t *testing.T) {
	testsuites.Driver(t, newDriverConstructor)
}

func BenchmarkInMemoryDriverSuite(b *testing.B) {
	testsuites.BenchDriver(b, newDriverConstructor)
}
