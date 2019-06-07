# Benchmarking Protocol Buffers
The following benchark was executed on a buffer of size 50 MB

| Test                        | Time      | Number of iterations | encoded buffer size |   |
|-----------------------------|-----------|----------------------|-------------|---|
| Protobuf                    | 282,841,146 | 5                    | 200 MB      |   |
| Protobuf (fixed32 encoding) | 110,602,110 | 10                   | 200 MB      |   |
| Protobuf with diff encoding | 243,161,755 | 5                    | 50 MB      |   |
| Protobuf with diff encoding (fixed 32) | 208,219,150 | 5                    | 50 MB      |   |
| Manual                      |  48,978,310 | 30                   | 200 MB      |   |

The following test does not contain source code for benchmarking protobuf with fixed32 encoding. It was done manually.
```go
func BenchmarkProtocolBuffers(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	keyCount := int32(5 * 1e7)
	temp := make([]uint32, keyCount)
	for i := 0; i < len(temp); i++ {
		temp[i] = uint32(i) + uint32(rand.Int31n(keyCount*4)) // keyCount * 4 to reduce collision (we want distinct offsets)
	}

	sort.Slice(temp, func(i, j int) bool { return temp[i] < temp[j] })

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
		diffTemp := make([]uint32, len(temp))
		diffTemp[0] = temp[0]
		for i := 1; i < len(temp); i++ {
			diffTemp[i] = temp[i] - temp[i-1]
		}
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
Output
```
go test -bench=BenchmarkProtocolBuffers -run=^$

goos: linux
goarch: amd64
pkg: github.com/dgraph-io/badger/table
BenchmarkProtocolBuffers/proto-16         	       5	 282841146 ns/op
--- BENCH: BenchmarkProtocolBuffers/proto-16
    table_test.go:793: buf length with proto 200 MB
    table_test.go:793: buf length with proto 200 MB
    table_test.go:793: buf length with proto 200 MB
BenchmarkProtocolBuffers/proto_with_offset_diff-16         	       5	 243161755 ns/op
--- BENCH: BenchmarkProtocolBuffers/proto_with_offset_diff-16
    table_test.go:814: buf length proto with diff 50 MB
    table_test.go:814: buf length proto with diff 50 MB
    table_test.go:814: buf length proto with diff 50 MB
BenchmarkProtocolBuffers/manual-16                         	      30	  48978310 ns/op
--- BENCH: BenchmarkProtocolBuffers/manual-16
    table_test.go:836: buf length manual 200 MB
    table_test.go:836: buf length manual 200 MB
PASS
ok  	github.com/dgraph-io/badger/table	53.802s

```