package bytemap

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	m = map[string]interface{}{
		"bool":    true,
		"byte":    byte(math.MaxUint8),
		"uint16":  uint16(math.MaxUint16),
		"uint32":  uint32(math.MaxUint32),
		"uint64":  uint64(math.MaxUint64),
		"uint":    uint(math.MaxUint64),
		"int8":    int8(math.MaxInt8),
		"int16":   int16(math.MaxInt16),
		"int32":   int32(math.MaxInt32),
		"int64":   int64(math.MaxInt64),
		"int":     math.MaxInt64,
		"float32": float32(math.MaxFloat32),
		"float64": float64(math.MaxFloat64),
		"string":  "Hello World",
		"time":    time.Now(),
		"nil":     nil,
	}

	sliceKeys = []string{"int16", "aunknown", "byte", "nil", "string"}
)

func init() {

}

func TestGet(t *testing.T) {
	bm := New(m)
	mbytes := map[string][]byte{}
	for key, value := range m {
		b := make([]byte, 100)
		_, n := encodeValue(b, value)
		b = b[:n]
		if len(b) == 0 {
			b = nil
		}
		mbytes[key] = b
	}

	for key, value := range m {
		assert.Equal(t, value, bm.Get(key))
		assert.EqualValues(t, mbytes[key], bm.GetBytes(key), fmt.Sprint(value))
	}
	assert.Nil(t, bm.Get("unspecified"))
}

func TestGetEmpty(t *testing.T) {
	bm := ByteMap(nil)
	assert.Nil(t, bm.Get("unspecified"))
}

func TestAsMap(t *testing.T) {
	m2 := New(m).AsMap()
	if assert.Equal(t, len(m), len(m2)) {
		for key, value := range m {
			assert.Equal(t, value, m2[key])
		}
	}
}

func TestAsMapEmpty(t *testing.T) {
	bm := ByteMap(nil)
	assert.Empty(t, bm.AsMap())
}

func TestFromSortedKeysAndValues(t *testing.T) {
	var keys []string
	var values []interface{}
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		values = append(values, m[key])
	}
	bm1 := New(m)
	bm2 := FromSortedKeysAndValues(keys, values)
	assert.EqualValues(t, bm1, bm2)
}

func TestFromSortedKeysAndFloats(t *testing.T) {
	m := map[string]interface{}{
		"a": float64(6.54),
		"b": float64(-72.32),
	}
	var keys []string
	var values []float64
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		values = append(values, m[key].(float64))
	}
	bm1 := New(m)
	bm2 := FromSortedKeysAndFloats(keys, values)
	assert.EqualValues(t, bm1, bm2)
}

func TestNilOnly(t *testing.T) {
	m2 := map[string]interface{}{
		"nil": nil,
	}
	bm := New(m2)
	assert.Nil(t, bm.Get("nil"))
	assert.Nil(t, bm.Get("unspecified"))
}

func TestSlice(t *testing.T) {
	bm := New(m)
	bm2 := bm.Slice(sliceKeys...)
	assert.True(t, len(bm2) < len(bm))
	for _, key := range sliceKeys {
		if "aunknown" == key {
			assert.Nil(t, bm2.Get(key))
		} else {
			assert.Equal(t, m[key], bm2.Get(key))
		}
	}
}

func TestSliceEmpty(t *testing.T) {
	bm := ByteMap(nil)
	assert.Empty(t, bm.Slice("unspecified").AsMap())
}

func BenchmarkByteMapAllKeys(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bm := New(m)
		for key := range m {
			bm.Get(key)
		}
	}
}

func BenchmarkByteMapOneKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bm := New(m)
		bm.Get("string")
	}
}

func BenchmarkByteSlice(b *testing.B) {
	bm := New(m)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.Slice(sliceKeys...)
	}
}

func BenchmarkMsgPackAllKeys(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b, _ := msgpack.Marshal(m)
		m2 := make(map[string]interface{}, 0)
		msgpack.Unmarshal(b, &m2)
		for key := range m {
			_ = m2[key]
		}
	}
}

func BenchmarkMsgPackOneKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b, _ := msgpack.Marshal(m)
		dec := msgpack.NewDecoder(bytes.NewReader(b))
		dec.Query("string")
	}
}

func BenchmarkMsgPackSlice(b *testing.B) {
	sliceKeysMap := make(map[string]bool, len(sliceKeys))
	for _, key := range sliceKeys {
		sliceKeysMap[key] = true
	}
	p, _ := msgpack.Marshal(m)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m2 := make(map[string]interface{}, 0)
		msgpack.Unmarshal(p, &m2)
		for key := range m2 {
			if !sliceKeysMap[key] {
				delete(m2, key)
			}
		}
		msgpack.Marshal(m2)
	}
}
