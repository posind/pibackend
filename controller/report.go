package controller

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func GetYesterdayDistincWAGroup(respw http.ResponseWriter, req *http.Request) {
	filter := bson.M{"_id": report.Yesterday()}
	res, err := atdb.GetAllDistinctDoc(config.Mongoconn, filter, "project.wagroupid", "pushrepo")
	if err != nil {
		var resp model.Response
		resp.Info = "Gagal Query Distincs project.wagroupid"
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	helper.WriteJSON(respw, http.StatusOK, res)
}

func GetReportHariIni(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	//kirim report ke group
	dt := &model.TextMessage{
		To:       "6281313112053-1492882006",
		IsGroup:  true,
		Messages: report.GetDataRepoMasukHarian(config.Mongoconn) + "\n" + report.GetDataLaporanMasukHarian(config.Mongoconn),
	}
	resp, err := helper.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	helper.WriteJSON(respw, http.StatusOK, resp)
}
