package geo_skeleton

import (
	"bytes"
	"compress/flate"
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"time"
)

const _letters string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewUUID generates and returns url friendly uuid.
// Source: http://play.golang.org/p/4FkNSiUDMg.
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(crand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	// return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
	return fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// NewUUID2 generates and returns a uuid.
func NewUUID2() (string, error) {
	b := make([]byte, 16)
	n, err := io.ReadFull(crand.Reader, b)
	if n != len(b) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	b[8] = b[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	b[6] = b[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// // CompressByte
// func CompressByte(src []byte) []byte {
// 	compressedData := new(bytes.Buffer)
// 	Compress(src, compressedData, 9)
// 	return compressedData.Bytes()
// }

// // DecompressByte
// func DecompressByte(src []byte) []byte {
// 	compressedData := bytes.NewBuffer(src)
// 	deCompressedData := new(bytes.Buffer)
// 	Decompress(compressedData, deCompressedData)
// 	return deCompressedData.Bytes()
// }

// // Compress
// func Compress(src []byte, dest io.Writer, level int) {
// 	compressor, _ := flate.NewWriter(dest, level)
// 	compressor.Write(src)
// 	compressor.Close()
// }

// // Decompress
// func Decompress(src io.Reader, dest io.Writer) {
// 	decompressor := flate.NewReader(src)
// 	io.Copy(dest, decompressor)
// 	decompressor.Close()
// }
