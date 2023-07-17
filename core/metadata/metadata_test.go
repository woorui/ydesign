package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD(t *testing.T) {
	var md = MD{}

	err := md.Set("abc", "def")
	assert.NoError(t, err)

	err = md.Set("a:v", "vvvv")
	assert.Error(t, ErrMDKeyOrValue, err)

	md2 := md.Clone()
	assert.Equal(t, md, md2)

	val, ok := md.Get("abc")
	assert.True(t, ok)
	assert.Equal(t, "def", val)

	data, err := md.Encode()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x61, 0x62, 0x63, 0x3a, 0x64, 0x65, 0x66}, data)

	var md3 = MD{}

	err = md3.Decode(data)
	assert.NoError(t, err)
	assert.Equal(t, md, md3)

	md.Del("abc")
	val, ok = md.Get("abc")
	assert.False(t, ok)
	assert.Equal(t, "", val)

}

func TestDecode(t *testing.T) {
	// Test case 1: Decoding valid data with multiple lines
	md := make(MD)
	err := md.Decode([]byte("key1:value1\nkey2:value2"))
	if err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}
	if len(md) != 2 {
		t.Errorf("Expected 2 key-value pairs, but got %d", len(md))
	}

	// Test case 2: Decoding valid data with single line
	md = make(MD)
	err = md.Decode([]byte("key:value"))
	if err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}
	if len(md) != 1 {
		t.Errorf("Expected 1 key-value pair, but got %d", len(md))
	}

	// Test case 3: Decoding invalid data with missing colon
	md = make(MD)
	err = md.Decode([]byte("keyvalue"))
	if err != ErrInvalidFormat {
		t.Errorf("Expected ErrInvalidFormat, but got %v", err)
	}
	if len(md) != 0 {
		t.Errorf("Expected 0 key-value pairs, but got %d", len(md))
	}

	// Test case 4: Decoding valid data with empty key and value
	md = make(MD)
	err = md.Decode([]byte(":"))
	if err != nil {
		t.Errorf("Expected nil error, but got %v", err)
	}
	if len(md) != 1 {
		t.Errorf("Expected 1 key-value pair, but got %d", len(md))
	}
	if md[""] != "" {
		t.Errorf("Expected empty key-value pair, but got %s:%s", "", md[""])
	}
}
