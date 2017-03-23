package geo_skeleton

import "github.com/sjsafranek/SkeletonDB"

// Database struct for application.
type Database struct {
	Table            string
	File             string
	commit_log_queue chan string
	Precision        int
	DB               skeleton.Database
}
