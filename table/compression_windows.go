// +build windows

/*
 * Copyright 2019 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package table

import (
	"github.com/dgraph-io/badger/options"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
)

var ErrUnsupported = errors.New("Compression algorithm not supported on this platform")

// compressData compresses the given data.
func (b *Builder) compressData(data []byte) ([]byte, error) {
	switch b.opt.Compression {
	case options.None:
		return data, nil
	case options.Snappy:
		return snappy.Encode(nil, data), nil
	case options.ZSTD:
		return nil, ErrUnsupported
	}
	return nil, errors.New("Unsupported compression type")
}

// decompressData decompresses the given data.
func (t *Table) decompressData(data []byte) ([]byte, error) {
	switch t.opt.Compression {
	case options.None:
		return data, nil
	case options.Snappy:
		return snappy.Decode(nil, data)
	case options.ZSTD:
		return nil, ErrUnsupported
	}
	return nil, errors.New("Unsupported compression type")
}
