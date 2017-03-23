package geo_skeleton

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

import (
	"./utils"
)

import (
	"github.com/paulmach/go.geojson"
	"github.com/sjsafranek/SkeletonDB"
)

// https://gist.github.com/DavidVaini/10308388
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func RoundToPrecision(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}

// DB application Database
var (
	COMMIT_LOG_FILE string = "geo_skeleton_commit.log"
)

// LayerCache keeps track of Database's loaded geojson layers
type LayerCache struct {
	Geojson *geojson.FeatureCollection
	Time    time.Time
}

func NewGeoSkeletonDB(db_file string) Database {
	var geoDb = Database{
		File:  db_file,
		Table: "GeoJsonDatasources",
		DB:    skeleton.Database{File: db_file}}
	geoDb.Init()
	return geoDb
}

// Database strust for application.
type Database struct {
	Table            string
	File             string
	Cache            map[string]*LayerCache
	guard            sync.RWMutex
	commit_log_queue chan string
	Precision        int
	DB               skeleton.Database
}

func (self Database) Init() {

	// Set initial data precision
	self.Precision = 8
	// Start db caching
	m := make(map[string]*LayerCache)
	self.Cache = m
	go self.cacheManager()
	// start commit log
	go self.startCommitLog()

	// default table
	if "" == self.Table {
		self.Table = "GeoJSONLayers"
	}

	conn := self.DB.Connect()
	defer conn.Close()

	err := self.DB.CreateTable(conn, self.Table)
	if nil != err {
		panic(err)
	}
}

// Starts Database commit log
func (self *Database) startCommitLog() {
	self.commit_log_queue = make(chan string, 10000)
	// open file to write database commit log
	COMMIT_LOG, err := os.OpenFile(COMMIT_LOG_FILE, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
	}
	defer COMMIT_LOG.Close()
	// read from chan and write to file
	for {
		if len(self.commit_log_queue) > 0 {
			line := <-self.commit_log_queue
			if _, err := COMMIT_LOG.WriteString(line + "\n"); err != nil {
				panic(err)
			}
		} else {
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

// CommitQueueLength returns length of database commit_log_queue
// @returns int
func (self *Database) CommitQueueLength() int {
	return len(self.commit_log_queue)
}

// NewLayer creates new datasource layer
// @returns string - datasource id
// @returns Error
// TODO: RENAME TO NewDatasource
func (self *Database) NewLayer() (string, error) {
	// create geojson
	datasource_id, _ := utils.NewUUID()
	geojs := geojson.NewFeatureCollection()
	// convert to bytes
	value, err := geojs.MarshalJSON()
	if err != nil {
		return "", nil
	}
	self.commit_log_queue <- `{"method": "create_datasource", "data": { "datasource": "` + datasource_id + `", "layer": ` + string(value) + `}}`
	// Insert layer into database
	err = self.DB.Insert(self.Table, datasource_id, value)
	if err != nil {
		panic(err)
	}
	return datasource_id, err
}

// InsertLayer inserts layer into database
// @param datasource {string}
// @param geojs {Geojson}
// @returns Error
func (self *Database) InsertLayer(datasource_id string, geojs *geojson.FeatureCollection) error {
	// Update caching layer
	if v, ok := self.Cache[datasource_id]; ok {
		self.guard.Lock()
		v.Geojson = geojs
		v.Time = time.Now()
		self.guard.Unlock()
	} else {
		pgc := &LayerCache{Geojson: geojs, Time: time.Now()}
		self.Cache[datasource_id] = pgc
	}
	// convert to bytes
	value, err := geojs.MarshalJSON()
	if err != nil {
		return err
	}
	err = self.DB.Insert(self.Table, datasource_id, value)
	if err != nil {
		panic(err)
	}

	return err
}

// GetLayer returns layer from database
// @param datasource {string}
// @returns Geojson
// @returns Error
func (self *Database) GetLayer(datasource_id string) (*geojson.FeatureCollection, error) {
	// Caching layer
	if v, ok := self.Cache[datasource_id]; ok {
		self.guard.RLock()
		v.Time = time.Now()
		self.guard.RUnlock()
		return v.Geojson, nil
	}
	// If cache ds not found get from database
	val, err := self.DB.Select(self.Table, datasource_id)
	if err != nil {
		return nil, err
	}
	if "" == string(val) {
		return nil, fmt.Errorf("Datasource not found")
	}
	// Read to struct
	geojs, err := geojson.UnmarshalFeatureCollection(val)
	if err != nil {
		return geojs, err
	}
	// Store page in memory cache
	pgc := &LayerCache{Geojson: geojs, Time: time.Now()}
	self.Cache[datasource_id] = pgc
	return geojs, nil
}

// DeleteLayer deletes layer from database
// @param datasource {string}
// @returns Error
func (self *Database) DeleteLayer(datasource_id string) error {
	self.commit_log_queue <- `{"method": "delete_layer", "data": { "datasource": "` + datasource_id + `"}}`

	err := self.DB.Remove(datasource_id, self.Table)

	self.guard.Lock()
	delete(self.Cache, datasource_id)
	self.guard.Unlock()
	return err
}

func (self *Database) normalizeGeometry(feat *geojson.Feature) (*geojson.Feature, error) {
	// FIT TO 7 - 8 DECIMAL PLACES OF PRECISION
	if nil == feat.Geometry {
		return nil, fmt.Errorf("Feature has no geometry!")
	}

	switch feat.Geometry.Type {

	case geojson.GeometryPoint:
		// []float64
		feat.Geometry.Point[0] = RoundToPrecision(feat.Geometry.Point[0], self.Precision)
		feat.Geometry.Point[1] = RoundToPrecision(feat.Geometry.Point[1], self.Precision)

	case geojson.GeometryMultiPoint:
		// [][]float64
		for i := range feat.Geometry.MultiPoint {
			for j := range feat.Geometry.MultiPoint[i] {
				feat.Geometry.MultiPoint[i][j] = RoundToPrecision(feat.Geometry.MultiPoint[i][j], self.Precision)
			}
		}

	case geojson.GeometryLineString:
		// [][]float64
		for i := range feat.Geometry.LineString {
			for j := range feat.Geometry.LineString[i] {
				feat.Geometry.LineString[i][j] = RoundToPrecision(feat.Geometry.LineString[i][j], self.Precision)
			}
		}

	case geojson.GeometryMultiLineString:
		// [][][]float64
		for i := range feat.Geometry.MultiLineString {
			for j := range feat.Geometry.MultiLineString[i] {
				for k := range feat.Geometry.MultiLineString[i][j] {
					feat.Geometry.MultiLineString[i][j][k] = RoundToPrecision(feat.Geometry.MultiLineString[i][j][k], self.Precision)
				}
			}
		}

	case geojson.GeometryPolygon:
		// [][][]float64
		for i := range feat.Geometry.Polygon {
			for j := range feat.Geometry.Polygon[i] {
				for k := range feat.Geometry.Polygon[i][j] {
					feat.Geometry.Polygon[i][j][k] = RoundToPrecision(feat.Geometry.Polygon[i][j][k], self.Precision)
				}
			}
		}

	case geojson.GeometryMultiPolygon:
		// [][][][]float64
		for i := range feat.Geometry.MultiPolygon {
			log.Printf("%v\n", feat.Geometry.MultiPolygon[i])
		}

	}

	/*
		//case GeometryCollection:
		//	geo.Geometries = g.Geometries
		//	// log.Printf("%v\n", feat.Geometry.Geometries)

	*/
	return feat, nil
}

func (self *Database) normalizeProperties(feat *geojson.Feature, featCollection *geojson.FeatureCollection) *geojson.Feature {

	// check if nil map
	if nil == feat.Properties {
		feat.Properties = make(map[string]interface{})
	}

	if 0 == len(featCollection.Features) {
		return feat
	}
	// Standardize properties for new feature
	for j := range featCollection.Features[0].Properties {
		if _, ok := feat.Properties[j]; !ok {
			feat.Properties[j] = ""
		}
	}

	// Standardize properties for existing features
	for i := range featCollection.Features {
		for j := range feat.Properties {
			if _, ok := featCollection.Features[i].Properties[j]; !ok {
				featCollection.Features[i].Properties[j] = ""
			}
		}
	}

	return feat
}

// InsertFeature adds feature to layer. Updates layer in Database
// @param datasource {string}
// @param feat {Geojson Feature}
// @returns Error
func (self *Database) InsertFeature(datasource_id string, feat *geojson.Feature) error {

	if nil == feat {
		return fmt.Errorf("feature value is <nil>!")
	}

	// Get layer from database
	featCollection, err := self.GetLayer(datasource_id)
	if err != nil {
		return err
	}

	// Apply required columns
	now := time.Now().Unix()

	// check if nil map
	if nil == feat.Properties {
		feat.Properties = make(map[string]interface{})
	}

	feat.Properties["is_active"] = true
	feat.Properties["is_deleted"] = false
	feat.Properties["date_created"] = now
	feat.Properties["date_modified"] = now
	feat.Properties["geo_id"] = fmt.Sprintf("%v", now)

	feat, err = self.normalizeGeometry(feat)
	if nil != err {
		return err
	}

	feat = self.normalizeProperties(feat, featCollection)

	// Write to commit log
	value, err := feat.MarshalJSON()
	if err != nil {
		return err
	}
	self.commit_log_queue <- `{"method": "insert_feature", "data": { "datasource": "` + datasource_id + `", "feature": ` + string(value) + `}}`

	// Add new feature to layer
	featCollection.AddFeature(feat)

	// insert layer
	err = self.InsertLayer(datasource_id, featCollection)
	if err != nil {
		panic(err)
	}
	return err
}

// EditFeature Edits feature in layer. Updates layer in Database
// @param datasource {string}
// @param geo_id {string}
// @param feat {Geojson Feature}
// @returns Error
func (self *Database) EditFeature(datasource_id string, geo_id string, feat *geojson.Feature) error {

	// Get layer from database
	featCollection, err := self.GetLayer(datasource_id)
	if err != nil {
		return err
	}

	feature_exists := false

	for i := range featCollection.Features {
		if geo_id == fmt.Sprintf("%v", featCollection.Features[i].Properties["geo_id"]) {

			now := time.Now().Unix()
			feat.Properties["date_modified"] = now

			feat, err = self.normalizeGeometry(feat)
			if nil != err {
				return err
			}

			feat = self.normalizeProperties(feat, featCollection)
			featCollection.Features[i] = feat
			// Write to commit log
			value, err := feat.MarshalJSON()
			if err != nil {
				return err
			}
			self.commit_log_queue <- `{"method": "edit_feature", "data": { "datasource": "` + datasource_id + `", "geo_id": "` + geo_id + `", "feature": ` + string(value) + `}}`
			feature_exists = true
		}
	}

	if !feature_exists {
		return fmt.Errorf("feature not found!")
	}

	// insert layer
	err = self.InsertLayer(datasource_id, featCollection)
	if err != nil {
		panic(err)
	}
	return err
}

// cacheManager for Database. Stores layers in memory.
//		Unloads layers older than 90 sec
//		When empty --> 60 sec timer
//		When items in cache --> 15 sec timer
func (self *Database) cacheManager() {
	for {
		n := float64(len(self.Cache))
		if n != 0 {
			for key := range self.Cache {
				// CHECK AVAILABLE SYSTEM MEMORY
				f := float64(len(self.Cache[key].Geojson.Features))
				limit := (300.0 - (f * (f * 0.25))) - (n * 2.0)
				if limit < 0.0 {
					limit = 10.0
				}
				if time.Since(self.Cache[key].Time).Seconds() > limit {
					self.guard.Lock()
					delete(self.Cache, key)
					self.guard.Unlock()
				}
			}
		}
		time.Sleep(10000 * time.Millisecond)
	}
}
