package cache_test

import (
	"crypto/sha256"
	"testing"
	"testing/quick"

	"github.com/opengapps/package-api/internal/pkg/cache"

	"github.com/stretchr/testify/assert"
)

func TestNewKey_Quick(t *testing.T) {
	f := func(x []byte) bool {
		key := cache.NewKey(x)
		hash := sha256.Sum256(x)
		k, ok := key.([32]byte)
		return ok && k == hash
	}

	err := quick.Check(f, nil)

	assert.NoError(t, err)
}
