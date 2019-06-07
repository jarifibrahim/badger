# Benchmarking Protocol Buffers
The following benchark was executed on a buffer of size 50 MB
| Test                        | Time      | Number of iterations | Buffer size |   |
|-----------------------------|-----------|----------------------|-------------|---|
| Protobuf                    | 297518315 | 5                    | 198 MB      |   |
| Protobuf (fixed32 encoding) | 111602110 | 10                   | 200 MB      |   |
| Protobuf with diff encoding | 441777861 | 3                    | 136 MB      |   |
| Manual                      |  53262966 | 30                   | 200 MB      |   |

```go
func BenchmarkProtocolBuffers(b *testing.B) {
	temp := make([]uint32, 5*1e7)
	for i := 0; i < len(temp); i++ {
		temp[i] = uint32(i) + uint32(rand.Intn(20))
	}

	b.Run("proto", func(b *testing.B) {
		m := pb.BlockMeta{
			EntryOffsets: temp,
		}
		mBuf, err := m.Marshal()
		b.Logf("proto %s", humanize.Bytes(uint64(len(mBuf))))
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
		b.Logf("proto with diff %s", humanize.Bytes(uint64(len(mBuf))))
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
		b.Logf("manual %s", humanize.Bytes(uint64(len(ebuf))))
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
Running tool: /usr/bin/go test -benchmem -run=^$ github.com/dgraph-io/badger/table -bench ^(BenchmarkReadRandom)$

goos: linux
goarch: amd64
pkg: github.com/dgraph-io/badger/table
BenchmarkReadRandom/proto-16         	       5	 297518315 ns/op	200008064 B/op	       5 allocs/op
--- BENCH: BenchmarkReadRandom/proto-16
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:789: proto 198 MB
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:789: proto 198 MB
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:789: proto 198 MB
BenchmarkReadRandom/proto_with_offset_diff-16         	       3	 441777861 ns/op	200008064 B/op	       5 allocs/op
--- BENCH: BenchmarkReadRandom/proto_with_offset_diff-16
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:810: proto with diff 136 MB
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:810: proto with diff 136 MB
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:810: proto with diff 136 MB
BenchmarkReadRandom/manual-16                         	      30	  53262966 ns/op	       0 B/op	       0 allocs/op
--- BENCH: BenchmarkReadRandom/manual-16
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:832: manual 200 MB
    /home/ibrahim/Projects/go/src/github.com/dgraph-io/badger/table/table_test.go:832: manual 200 MB
PASS
ok  	github.com/dgraph-io/badger/table	47.278s
Success: Benchmarks passed.

```