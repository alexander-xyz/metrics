package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignIsStable(t *testing.T) {
	data := []byte(`{"id":"Alloc","type":"gauge","value":1.5}`)

	assert.Equal(t, Sign(data, "key"), Sign(data, "key"))
	assert.NotEqual(t, Sign(data, "key"), Sign(data, "other"))
	assert.NotEqual(t, Sign(data, "key"), Sign([]byte("other"), "key"))
}

func TestValid(t *testing.T) {
	data := []byte("payload")
	sum := Sign(data, "key")

	assert.True(t, Valid(data, "key", sum))
	assert.False(t, Valid(data, "other", sum), "чужой ключ")
	assert.False(t, Valid([]byte("changed"), "key", sum), "изменённые данные")
	assert.False(t, Valid(data, "key", "не hex"), "некорректная подпись")
	assert.False(t, Valid(data, "key", ""))
}
