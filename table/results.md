# Benchmarking Protocol Buffers
The following benchark was executed on a buffer of size 40 MB

| Test                                           | Time        | # of iterations | encoded buffer size |
|------------------------------------------------|-------------|-----------------|---------------------|
| Protobuf                                       | 229,493,210 | 5               | 160 MB              |
| Protobuf (fixed32 encoding)                    | 84,007,882  | 20              | 160 MB              |
| Protobuf with diff encoding                    | 193,642,099 | 10              | 40 MB               |
| Protobuf with diff encoding (fixed32 encoding) | 165,258,664 | 10              | 160 MB              |
| GroupVarint with diff encoding                 | 121,073,602 | 10              | 50 MB               |
| Manual                                         | 41,154,421  | 30              | 160 MB              |

The following test does not contain source code for benchmarking protobuf with fixed32 encoding. It was done manually.
```go
func BenchmarkProtocolBuffers(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	keyCount := int32(4 * 1e7) // should be in multiple of 4 for groupvarint to work
	temp := make([]uint32, keyCount)
	for i := 0; i < len(temp); i++ {
		temp[i] = uint32(i) + uint32(rand.Int31n(keyCount*4)) // keyCount * 4 to reduce collision (we want distinct offsets)
	}
	// Sort temp array so that we get diff in increasing order
	sort.Slice(temp, func(i, j int) bool { return temp[i] < temp[j] })
	diffTemp := make([]uint32, len(temp))
	diffTemp[0] = temp[0]
	for i := 1; i < len(temp); i++ {
		diffTemp[i] = temp[i] - temp[i-1]
	}

	b.Run("diff with groupvarint", func(b *testing.B) {
		// Encode data
		data := make([]byte, 0)
		for i := 0; i <= len(diffTemp)-4; i += 4 {
			buf := make([]byte, 17)
			tempdata := groupvarint.Encode4(buf, diffTemp[i:i+4])
			data = append(data, tempdata...)
		}
		deltas := append(data, 0, 0, 0)
		b.Logf("diff with groupvarint %s", humanize.Bytes(uint64(len(deltas))))
		b.ResetTimer()
		buf := make([]uint32, 4)
		for j := 0; j < b.N; j++ {
			b.StopTimer()
			deltaCopy := make([]byte, len(deltas))
			entryOffsets := make([]uint32, 0)
			copy(deltaCopy, deltas)
			b.StartTimer()
			// decode data
			for len(deltaCopy) >= 5 {
				groupvarint.Decode4(buf, deltaCopy)
				deltaCopy = deltaCopy[groupvarint.BytesUsed[deltaCopy[0]]:]
				entryOffsets = append(entryOffsets, buf...)
			}
			b.StopTimer()
			require.Equal(b, diffTemp, entryOffsets)
			b.StartTimer()
		}
	})
	b.Run("proto", func(b *testing.B) {
		m := pb.BlockMeta{
			EntryOffsets: temp,
		}
		mBuf, err := m.Marshal()
		b.Logf("buf length with proto %s", humanize.Bytes(uint64(len(mBuf))))
		require.NoError(b, err)
		k := pb.BlockMeta{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k.Reset()
			require.NoError(b, k.Unmarshal(mBuf))
		}
		b.StopTimer()
		require.EqualValues(b, temp, k.EntryOffsets)
	})
	b.Run("proto with offset diff", func(b *testing.B) {
		m := pb.BlockMeta{
			EntryOffsets: diffTemp,
		}
		mBuf, err := m.Marshal()
		b.Logf("buf length proto with diff %s", humanize.Bytes(uint64(len(mBuf))))
		require.NoError(b, err)
		k := pb.BlockMeta{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			k.Reset()
			require.NoError(b, k.Unmarshal(mBuf))
			for j := 1; j < len(k.EntryOffsets); j++ {
				k.EntryOffsets[j] += k.EntryOffsets[j-1]
			}
		}
		b.StopTimer()
		require.EqualValues(b, temp, k.EntryOffsets)
	})
	b.Run("manual", func(b *testing.B) {
		ebuf := make([]byte, len(temp)*4)
		e1 := ebuf
		for i := len(temp) - 1; i >= 0; i-- {
			binary.BigEndian.PutUint32(e1[:4], temp[i])
			e1 = e1[4:]
		}
		entryOffsets := make([]uint32, len(temp))
		b.Logf("buf length manual %s", humanize.Bytes(uint64(len(ebuf))))
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			readPos := len(ebuf) - 4
			for i := 0; i < len(temp); i++ {
				entryOffsets[i] = binary.BigEndian.Uint32(ebuf[readPos : readPos+4])
				readPos -= 4
			}
		}
		b.StopTimer()
		require.EqualValues(b, temp, entryOffsets)
	})

}
```
Output (with uint32 protobuf)
```
goos: linux
goarch: amd64
pkg: github.com/dgraph-io/badger/table
BenchmarkProtocolBuffers/diff_with_groupvarint-16         	      10	 121073602 ns/op
--- BENCH: BenchmarkProtocolBuffers/diff_with_groupvarint-16
    table_test.go:804: diff with groupvarint 50 MB
    table_test.go:804: diff with groupvarint 50 MB
    table_test.go:804: diff with groupvarint 50 MB
BenchmarkProtocolBuffers/proto-16                         	       5	 229493210 ns/op
--- BENCH: BenchmarkProtocolBuffers/proto-16
    table_test.go:829: buf length with proto 160 MB
    table_test.go:829: buf length with proto 160 MB
BenchmarkProtocolBuffers/proto_with_offset_diff-16        	      10	 193642099 ns/op
--- BENCH: BenchmarkProtocolBuffers/proto_with_offset_diff-16
    table_test.go:845: buf length proto with diff 40 MB
    table_test.go:845: buf length proto with diff 40 MB
BenchmarkProtocolBuffers/manual-16                        	      30	  41154421 ns/op
--- BENCH: BenchmarkProtocolBuffers/manual-16
    table_test.go:867: buf length manual 160 MB
    table_test.go:867: buf length manual 160 MB
PASS
ok  	github.com/dgraph-io/badger/table	94.496s
```
output (with fixed32 protobuf)
```
goos: linux
goarch: amd64
pkg: github.com/dgraph-io/badger/table
BenchmarkProtocolBuffers/diff_with_groupvarint-16         	      10	 122372446 ns/op
--- BENCH: BenchmarkProtocolBuffers/diff_with_groupvarint-16
    table_test.go:804: diff with groupvarint 50 MB
    table_test.go:804: diff with groupvarint 50 MB
BenchmarkProtocolBuffers/proto-16                         	      20	  84007882 ns/op
--- BENCH: BenchmarkProtocolBuffers/proto-16
    table_test.go:829: buf length with proto 160 MB
    table_test.go:829: buf length with proto 160 MB
BenchmarkProtocolBuffers/proto_with_offset_diff-16        	      10	 165258664 ns/op
--- BENCH: BenchmarkProtocolBuffers/proto_with_offset_diff-16
    table_test.go:845: buf length proto with diff 160 MB
    table_test.go:845: buf length proto with diff 160 MB
BenchmarkProtocolBuffers/manual-16                        	      30	  39146657 ns/op
--- BENCH: BenchmarkProtocolBuffers/manual-16
    table_test.go:867: buf length manual 160 MB
    table_test.go:867: buf length manual 160 MB
PASS
ok  	github.com/dgraph-io/badger/table	75.465s
```