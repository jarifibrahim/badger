package y

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dustin/go-humanize"

	"github.com/stretchr/testify/require"
)

func TestReadNoalloc(t *testing.T) {
	n := NewNoAllocBuffer(5)
	var result []byte
	for i := 0; i < 4; i++ {
		res := []byte(fmt.Sprintf("%4d", i))
		result = append(result, res...)
		_, err := n.Write(res)
		require.NoError(t, err)
	}
	require.Equal(t, n.Bytes(), result)
}

func BenchmarkBuffer(b *testing.B) {
	count := int(1 * 1e5) // 100 KB * 10 == 1 MB
	bufSz := 1 << 10
	key := func(i int) []byte {
		return []byte(fmt.Sprintf("%10d", i))
	}
	totalSz := humanize.Bytes(uint64(count * 10))
	b.Run(fmt.Sprintf("NoAlloc-write-%s", totalSz), func(b *testing.B) {
		n := NewNoAllocBuffer(bufSz)
		fmt.Println("Page size:", humanize.Bytes(uint64(bufSz)))
		for i := 0; i < b.N; i++ {
			for j := 0; j < count; j++ {
				_, err := n.Write(key(j))
				require.NoError(b, err)
			}
		}
	})
	b.Run(fmt.Sprintf("Bytes Buffer-write-%s", totalSz), func(b *testing.B) {
		buf := bytes.Buffer{}
		buf.Grow(bufSz)
		for i := 0; i < b.N; i++ {
			for j := 0; j < count; j++ {
				_, err := buf.Write(key(j))
				require.NoError(b, err)
			}
		}
	})

	b.Run(fmt.Sprintf("NoAlloc-Bytes-%s", totalSz), func(b *testing.B) {
		n := NewNoAllocBuffer(1 << 10)
		for j := 0; j < count; j++ {
			_, err := n.Write(key(j))
			require.NoError(b, err)
		}
		for i := 0; i < b.N; i++ {
			require.Equal(b, count*10, len(n.Bytes()))
		}
	})
	b.Run(fmt.Sprintf("Bytes Buffer-Bytes-%s", totalSz), func(b *testing.B) {
		buf := bytes.Buffer{}
		buf.Grow(1 << 10)
		for j := 0; j < count; j++ {
			_, err := buf.Write(key(j))
			require.NoError(b, err)
		}
		for i := 0; i < b.N; i++ {
			require.Equal(b, count*10, len(buf.Bytes()))
		}
	})
}
