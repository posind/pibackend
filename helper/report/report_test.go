package report

import (
	"fmt"
	"os"
	"testing"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
)

var mongoinfo = atdb.DBInfo{
	DBString: os.Getenv("MONGODOMYID"),
	DBName:   "domyid",
}

var Mongoconn, ErrorMongoconn = atdb.MongoConnect(mongoinfo)

func TestGenerateReport(t *testing.T) {
	config.WAAPIToken = "v4.public.eyJhbGlhcyI6IiIsImV4cCI6IjIwMjQtMDgtMjlUMjA6MDA6MDBaIiwiaWF0IjoiMjAyNC0wNy0zMFQyMDowMDowMFoiLCJpZCI6IjYyODk1NjAxMDYwMDAwIiwibmJmIjoiMjAyNC0wNy0zMFQyMDowMDowMFoiffszbcuYZe40dHBOjcsC3IyXNBXbZuKniOlFbazHhRKOUYapjWqAfafLpRZ8JzfXfgaCzixLAfI8nRBSEgBPIgM"
	fmt.Println(mongoinfo.DBString)
	err := RekapMeetingKemarin(Mongoconn)
	fmt.Println(err)
	//fmt.Println(md)

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
