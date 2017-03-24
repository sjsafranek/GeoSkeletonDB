package geoskeleton

import (
	"fmt"
	"log"
	"os"
	"time"
)

import (
	"github.com/paulmach/go.geojson"
	"github.com/sjsafranek/DiffDB/diff_store"
	"github.com/sjsafranek/SkeletonDB"
)

var (
	// COMMIT_LOG_FILE database commit log file path
	COMMIT_LOG_FILE string = "geo_skeleton_commit.log"
)

const (
	// DEFAULT_PRECISION decimal decimal for storing latitude and longitude.
	DEFAULT_PRECISION      int    = 6 //8

	// DEFAULT_DATABASE_TABLE Database table used to store data.
	DEFAULT_DATABASE_TABLE string = "GeoJsonDatasources"
)

// Creates a GeoSkeletonDB
func NewGeoSkeletonDB(db_file string) Database {
	var geoDb = Database{
		File:  db_file,
		Table: DEFAULT_DATABASE_TABLE,
		DB:    skeleton.Database{File: db_file}}
	geoDb.Init()
	return geoDb
}

// Initiates database.
// Creates required database tables and starts commit log.
func (self *Database) Init() {

	// start commit log
	go self.StartCommitLog()

	self.DB.Init()

	conn := self.DB.Connect()
	defer conn.Close()

	err := self.DB.CreateTable(conn, self.getTable())
	if nil != err {
		panic(err)
	}

	err = self.DB.CreateTable(conn, "GeoTimeseriesData")
	if nil != err {
		panic(err)
	}
}

// Get database table
func (self *Database) getTable() string {
	if "" == self.Table {
		return DEFAULT_DATABASE_TABLE
	}
	return self.Table
}

// Get precision for rounding latitude longitude values.
func (self *Database) getPrecision() int {
	if 1 > self.Precision {
		return DEFAULT_PRECISION
	}
	return self.Precision
}

// Starts Database commit log.
func (self *Database) StartCommitLog() {
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

// CommitQueueLength returns length of database commit_log_queue.
func (self *Database) CommitQueueLength() int {
	return len(self.commit_log_queue)
}

// NewLayer creates new geojson layer.
// Writes new layer to database.
// TODO: RENAME TO NewDatasource
func (self *Database) NewLayer() (string, error) {
	// create geojson
	datasource_id, _ := NewUUID()
	geojs := geojson.NewFeatureCollection()
	// convert to bytes
	value, err := geojs.MarshalJSON()
	if err != nil {
		return "", nil
	}
	self.commit_log_queue <- `{"method": "create_datasource", "data": { "datasource": "` + datasource_id + `", "layer": ` + string(value) + `}}`
	// Insert layer into database
	err = self.DB.Insert(self.getTable(), datasource_id, value)
	if err != nil {
		panic(err)
	}
	return datasource_id, err
}

// InsertLayer inserts geojson layer into database
// TODO: Switch to timeseries datasource
func (self *Database) InsertLayer(datasource_id string, geojs *geojson.FeatureCollection) error {
	// convert to bytes
	value, err := geojs.MarshalJSON()
	if err != nil {
		return err
	}

	err = self.DB.Insert(self.getTable(), datasource_id, value)
	if err != nil {
		panic(err)
	}

	go self.UpdateTimeseriesDatasource(datasource_id, value)

	return err
}

// GetLayer returns geojson layer from database
// TODO: Switch to timeseries datasource
func (self *Database) GetLayer(datasource_id string) (*geojson.FeatureCollection, error) {
	val, err := self.DB.Select(self.getTable(), datasource_id)
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
	return geojs, nil
}

// GetLayers returns all datasource_ids from database
func (self *Database) GetLayers() ([]string, error) {
	val, err := self.DB.SelectAll(self.getTable())
	if err != nil {
		return nil, err
	}
	return val, nil
}

// DeleteLayer deletes geojson layer from database
func (self *Database) DeleteLayer(datasource_id string) error {
	self.commit_log_queue <- `{"method": "delete_layer", "data": { "datasource": "` + datasource_id + `"}}`
	err := self.DB.Remove(datasource_id, self.getTable())
	return err
}

// Normalizes geometroy to set decimal precision
func (self *Database) normalizeGeometry(feat *geojson.Feature) (*geojson.Feature, error) {
	// FIT TO 7 - 8 DECIMAL PLACES OF PRECISION
	if nil == feat.Geometry {
		return nil, fmt.Errorf("Feature has no geometry!")
	}

	precision := self.getPrecision()

	switch feat.Geometry.Type {

	case geojson.GeometryPoint:
		// []float64
		feat.Geometry.Point[0] = RoundToPrecision(feat.Geometry.Point[0], precision)
		feat.Geometry.Point[1] = RoundToPrecision(feat.Geometry.Point[1], precision)

	case geojson.GeometryMultiPoint:
		// [][]float64
		for i := range feat.Geometry.MultiPoint {
			for j := range feat.Geometry.MultiPoint[i] {
				feat.Geometry.MultiPoint[i][j] = RoundToPrecision(feat.Geometry.MultiPoint[i][j], precision)
			}
		}

	case geojson.GeometryLineString:
		// [][]float64
		for i := range feat.Geometry.LineString {
			for j := range feat.Geometry.LineString[i] {
				feat.Geometry.LineString[i][j] = RoundToPrecision(feat.Geometry.LineString[i][j], precision)
			}
		}

	case geojson.GeometryMultiLineString:
		// [][][]float64
		for i := range feat.Geometry.MultiLineString {
			for j := range feat.Geometry.MultiLineString[i] {
				for k := range feat.Geometry.MultiLineString[i][j] {
					feat.Geometry.MultiLineString[i][j][k] = RoundToPrecision(feat.Geometry.MultiLineString[i][j][k], precision)
				}
			}
		}

	case geojson.GeometryPolygon:
		// [][][]float64
		for i := range feat.Geometry.Polygon {
			for j := range feat.Geometry.Polygon[i] {
				for k := range feat.Geometry.Polygon[i][j] {
					feat.Geometry.Polygon[i][j][k] = RoundToPrecision(feat.Geometry.Polygon[i][j][k], precision)
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

// Normalizes properties within geojson layer using geojson feature.
// Normalizes properties within geojson feature using geojson layers.
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

// InsertFeature adds geojson feature to geojson layer.
// Updates layer in database
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

// EditFeature edits geojson feature within geojson layer.
// Updates geojson layer in Database
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

// InsertTimeseriesDatasource inserts timeseries geojson layer to database
func (self *Database) InsertTimeseriesDatasource(datasource_id string, ddata diff_store.DiffStore) error {
	// save to database
	enc, err := ddata.Encode()
	if nil != err {
		panic(err)
	}

	// not matching ?!?!
	ddata.Name = datasource_id

	err = self.DB.Insert("GeoTimeseriesData", datasource_id, enc)
	return err
}

// SelectTimeseriesDatasource selects timeseries geojson layer from database
func (self *Database) SelectTimeseriesDatasource(datasource_id string) (diff_store.DiffStore, error) {
	var ddata diff_store.DiffStore
	data, err := self.DB.Select("GeoTimeseriesData", datasource_id)
	ddata.Decode(data)
	return ddata, err
}

// UpdateTimeseriesDatasource updates timeseries geojson layer
// and saves to database.
func (self *Database) UpdateTimeseriesDatasource(datasource_id string, value []byte) error {
	// get diffstore record
	ddata, err := self.SelectTimeseriesDatasource(datasource_id)
	if nil != err {
		if err.Error() == "Not found" {
			// create new diffstore if key not found in database
			ddata = diff_store.NewDiffStore(datasource_id)
		} else {
			panic(err)
		}
	}

	// update diffstore
	update_value := string(value)
	ddata.Update(update_value)

	// write to database
	err = self.InsertTimeseriesDatasource(datasource_id, ddata)
	return err
}
