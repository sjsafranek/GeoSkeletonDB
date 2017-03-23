package main

import (
	"fmt"
)

import (
	"./geo_skeleton_db"
)

const (
	NAME   = "GeoSkeletonDB Client"
	BINARY = "geo_skeleton_db_cli"
)

func main() {

	db := geo_skeleton.NewGeoSkeletonDB("geo_bolt.db")
	fmt.Printf("%v\n", db)

}
