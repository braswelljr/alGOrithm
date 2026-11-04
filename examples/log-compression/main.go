// Command log-compression shrinks repetitive log text two ways: Huffman
// coding gives frequent characters short codes (as in DEFLATE, used by gzip,
// zip and PNG), and LZW replaces repeated phrases with dictionary codes (as
// in GIF and Unix compress).
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/compression"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/greedy"
)

func main() {
	logs := strings.Repeat("GET /api/users 200 12ms\nGET /api/orders 200 9ms\n", 20)
	h, err := greedy.HuffmanCoding(logs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("original: %d bytes\n", len(logs))
	fmt.Printf("huffman:  %d bytes (%.0f%% of original)\n", (h.Bits+7)/8, 100*float64(h.Bits)/float64(h.OriginalBits))
	codes, _ := compression.LZW([]byte(logs))
	lzwBytes := len(codes) * 12 / 8 // 12-bit codes, as classic LZW uses
	fmt.Printf("lzw:      %d codes, about %d bytes (%.0f%% of original)\n", len(codes), lzwBytes, 100*float64(lzwBytes)/float64(len(logs)))
	back, err := compression.LZWDecode(codes)
	fmt.Printf("lzw round trip exact: %v\n", err == nil && string(back) == logs)
}
