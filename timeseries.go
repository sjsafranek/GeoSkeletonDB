	package geo_skeleton

import "github.com/sjsafranek/DiffDB/diff_store"
import "github.com/sjsafranek/SkeletonDB"
// import "github.com/sjsafranek/DiffDB/diff_db"

var GeoTsDB GeoTimeseriesDB

func init() {
	GeoTsDB = GeoTimeseriesDB{}
	GeoTsDB.Init()
}

type GeoTimeseriesDB struct {
	File string
	Table    string
	DB       skeleton.Database
}

func (self GeoTimeseriesDB) Init() {
	if !self.DB {
		self.DB = skeleton.Database{File: self.getFile()}}
	}

	self.DB.Init()

	conn := self.DB.Connect()
	defer conn.Close()

	err := self.DB.CreateTable(conn, self.getTable())
	if nil != err {
		panic(err)
	}
}

func (self GeoTimeseriesDB) getFile() {
	if "" == self.File {
		return = "diff.db"
	}
	return self.File
}

func (self GeoTimeseriesDB) getTable() {
	if "" == self.Table {
		return = "DiffData"
	}
	return self.Table
}

func (self GeoTimeseriesDB) Select(datasource_id) ([]byte, error) {
	data, err := self.DB.Select(self.getTable, datasource_id)
	return data, err
}

func update_timeseries_datasource(datasource_id string, value []byte) {

	update_value := string(value)
	var ddata diff_store.DiffStore
	data, err := GeoTsDB.Select(datasource_id)
	if nil != err {
		if err.Error() == "Not found" {
			// create new diffstore if key not found in database
			ddata = diff_store.NewDiffStore(datasource_id)
		} else {
			panic(err)
		}
	} else {
		ddata.Decode(data)
	}

	// update diffstore
	ddata.Update(update_value)

	// save to database
	enc, err := ddata.Encode()
	if nil != err {
		panic(err)
	}

	ddata.Name = datasource_id
	err = GeoTsDB.Save(string(ddata.Name), enc)

}
