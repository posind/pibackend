package report

import (
	"os"
	"testing"

	"github.com/gocroot/helper"
	"github.com/gocroot/model"
)

var mongoinfo = model.DBInfo{
	DBString: os.Getenv("MONGODOMYID"),
	DBName:   "domyid",
}

var Mongoconn, ErrorMongoconn = helper.MongoConnect(mongoinfo)

func TestGenerateReport(t *testing.T) {
	results := GetDataRepoMasukKemarin(Mongoconn, "6281313112053-1492882006") // + "\n" + GetDataLaporanMasukKemarin(Mongoconn)
	print(results)

}
