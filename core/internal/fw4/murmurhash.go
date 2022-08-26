// Copyright 2022 Dynatrace LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fw4

// Adapted from OneAgent .NET 1.235.0.20220114-170639

const defaultSeed int64 = 0xe17a1465

func latin1MurmurHash2_64A(data string) int64 {
	return latin1MurmurHash2_64A_Custom(data, defaultSeed, 0, len(data))
}

func latin1MurmurHash2_64A_Custom(data string, seed int64, startIdx int, endIdx int) int64 {
	len := endIdx - startIdx

	var m int64 = -0x395b586ca42e166b
	var r int32 = 47

	var h int64 = seed ^ int64(len)*m

	var fullChunksEndIdx int = startIdx + (int)(len&0xFFFFFFF8)

	var idx int
	for idx = startIdx; idx < fullChunksEndIdx; idx += 8 {
		b0 := int64(data[idx] & 0xFF)
		b1 := int64(data[idx+1] & 0xFF)
		b2 := int64(data[idx+2] & 0xFF)
		b3 := int64(data[idx+3] & 0xFF)
		b4 := int64(data[idx+4] & 0xFF)
		b5 := int64(data[idx+5] & 0xFF)
		b6 := int64(data[idx+6] & 0xFF)
		b7 := int64(data[idx+7] & 0xFF)

		var bb0 int64 = b0 | (b4 << 32) // optimization for SIMD
		var bb1 int64 = b1 | (b5 << 32)
		var bb2 int64 = b2 | (b6 << 32)
		var bb3 int64 = b3 | (b7 << 32)

		var bbb0 int64 = bb0 | (bb2 << 16)
		var bbb1 int64 = bb1 | (bb3 << 16)

		var k int64 = bbb0 | (bbb1 << 8)

		k *= m
		k ^= int64(uint64(k) >> r)
		k *= m
		h ^= k
		h *= m
	}

	switch len & 0x7 {
	case 7:
		h ^= int64(uint64(data[idx+6]&0xFF) << 48)
		fallthrough
	case 6:
		h ^= int64(uint64(data[idx+5]&0xFF) << 40)
		fallthrough
	case 5:
		h ^= int64(uint64(data[idx+4]&0xFF) << 32)
		fallthrough
	case 4:
		h ^= int64(uint64(data[idx+3]&0xFF) << 24)
		fallthrough
	case 3:
		h ^= int64(uint64(data[idx+2]&0xFF) << 16)
		fallthrough
	case 2:
		h ^= int64(uint64(data[idx+1]&0xFF) << 8)
		fallthrough
	case 1:
		h ^= int64(data[idx] & 0xFF)
		h *= m
	}

	h ^= int64(uint64(h) >> r)
	h *= m
	h ^= int64(uint64(h) >> r)

	return h
}
