package report

import (
	"testing"

	"github.com/gocroot/helper"
	"github.com/gocroot/model"
)

var mongoinfo = model.DBInfo{
	DBString: "mongodb+srv://ulbi:k0dGfeYgAorMKDAz@cluster0.fvazjna.mongodb.net/",
	DBName:   "domyid",
}

var Mongoconn, ErrorMongoconn = helper.MongoConnect(mongoinfo)

func TestGenerateReport(t *testing.T) {
	results := GetDataRepoMasukHarian(Mongoconn) + "\n" + GetDataLaporanMasukHarian(Mongoconn)
	print(results)

}
