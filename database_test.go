package geoskeleton

import (
	"fmt"
	"testing"
)

// TestInitDB
func TestInitDB(t *testing.T) {

	db := NewGeoSkeletonDB("test.db")
	fmt.Printf("%v\n", db)
	// if 4 != len(str) {
	// 	t.Error("Length of value is not what was expected")
	// }
}
