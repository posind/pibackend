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
	//gid := "6281313112053-1492882006"
	gid := "6281312000300-1488324890"
	results, err := GenerateRekapMessageKemarinPerWAGroupID(Mongoconn, gid) // + "\n" + GetDataLaporanMasukKemarin(Mongoconn)
	if err != nil {
		print(err.Error())
	}

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
