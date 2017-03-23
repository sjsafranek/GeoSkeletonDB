package geo_skeleton

import "github.com/sjsafranek/DiffDB/diff_store"
import "github.com/sjsafranek/SkeletonDB"

type GeoTimeseriesDB struct {
	File string
	Table    string
	DB       skeleton.Database
}

func (self GeoTimeseriesDB) Init() {

	self.DB = skeleton.Database{File: string(self.getFile())}
	self.DB.Init()

	conn := self.DB.Connect()
	defer conn.Close()

	err := self.DB.CreateTable(conn, string(self.getTable()))
	if nil != err {
		panic(err)
	}
}

func (self GeoTimeseriesDB) getFile() string {
	if "" == self.File {
		return "geo_ts.db"
	}
	return self.File
}

func (self GeoTimeseriesDB) getTable() string {
	if "" == self.Table {
		return "GeoTimeseriesData"
	}
	return self.Table
}

func (self GeoTimeseriesDB) Insert(datasource_id string, enc []byte) (error) {
	err := self.DB.Insert(self.getTable(), datasource_id, enc)
	return err
}

func (self GeoTimeseriesDB) Select(datasource_id string) ([]byte, error) {
	data, err := self.DB.Select(self.getTable(), datasource_id)
	return data, err
}

func (self GeoTimeseriesDB) UpdateTimeseriesDatasource(datasource_id string, value []byte) {

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
	err = GeoTsDB.Insert(string(ddata.Name), enc)

}
