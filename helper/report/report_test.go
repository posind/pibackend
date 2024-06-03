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
	gid := "6281313112053-1492882006"
	results := GetDataRepoMasukHariIniPerWaGroupID(Mongoconn, gid) // + "\n" + GetDataLaporanMasukKemarin(Mongoconn)
	print(results)

}

/* func TestGenerateReportLayanan(t *testing.T) {
	gid := "6281313112053-1492882006"
	results := GetDataLaporanMasukHariini(Mongoconn, gid) //GetDataLaporanMasukHarian
	print(results)

}

func TestGenerateReportLay(t *testing.T) {
	//gid := "6281313112053-1492882006"
	results := GetDataLaporanMasukHarian(Mongoconn) //GetDataLaporanMasukHarian
	print(results)

} */
