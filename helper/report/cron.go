package report

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func RekapTengahMalam(respw http.ResponseWriter, req *http.Request) {
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
		dt := &model.TextMessage{
			To:       groupID,
			IsGroup:  true,
			Messages: GetDataRepoMasukHariIni(config.Mongoconn, groupID) + "\n" + GetDataLaporanMasukHariini(config.Mongoconn, groupID),
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
