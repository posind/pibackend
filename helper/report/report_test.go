package report

import (
	"fmt"
	"os"
	"testing"

	"github.com/gocroot/helper/atdb"
)

var mongoinfo = atdb.DBInfo{
	DBString: os.Getenv("MONGODOMYID"),
	DBName:   "domyid",
}

var Mongoconn, ErrorMongoconn = atdb.MongoConnect(mongoinfo)

func TestGenerateReport(t *testing.T) {
	fmt.Println(mongoinfo.DBString)
	RekapMeetingKemarin(Mongoconn, "lmsdesa")

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
