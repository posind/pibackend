package report

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func RekapPagiHari(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	filter := bson.M{"_id": TodayFilter()}
	wagroupidlist, err := atdb.GetAllDistinctDoc(config.Mongoconn, filter, "project.wagroupid", "pushrepo")
	if err != nil {
		resp.Info = "Gagal Query Distincs project.wagroupid"
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	for _, gid := range wagroupidlist { //iterasi di setiap wa group
		// Type assertion to convert any to string
		groupID, ok := gid.(string)
		if !ok {
			resp.Info = "wagroupid is not a string"
			resp.Response = "wagroupid is not a string"
			helper.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
		msg, err := GenerateRekapMessageKemarinPerWAGroupID(config.Mongoconn, groupID)
		if err != nil {
			resp.Info = "Gagal Membuat Rekapitulasi perhitungan per wa group id"
			resp.Response = err.Error()
			helper.WriteJSON(respw, http.StatusExpectationFailed, resp)
			return
		}
		dt := &model.TextMessage{
			To:       groupID,
			IsGroup:  true,
			Messages: msg,
		}
		resp, err = helper.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
		if err != nil {
			resp.Info = "Tidak berhak"
			resp.Response = err.Error()
			helper.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
	}

}
