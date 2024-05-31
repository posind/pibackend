package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func GetHome(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	resp.Response = helper.GetIPaddress()
	helper.WriteJSON(respw, http.StatusOK, resp)
}

func PostInboxNomor(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	var msg model.IteungMessage
	httpstatus := http.StatusUnauthorized
	resp.Response = "Wrong Secret"
	waphonenumber := helper.GetParam(req)
	prof, err := helper.GetAppProfile(waphonenumber, config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		httpstatus = http.StatusServiceUnavailable
	}
	if helper.GetSecretFromHeader(req) == prof.Secret {
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			resp.Response = err.Error()
		} else {
			resp, err = helper.WebHook(prof.QRKeyword, waphonenumber, config.WAAPIQRLogin, config.WAAPIMessage, msg, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
			}
		}
	}
	helper.WriteJSON(respw, httpstatus, resp)
}

func GetNewToken(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	httpstatus := http.StatusServiceUnavailable
	//prof, err := helper.GetAppProfile(config.WAPhoneNumber, config.Mongoconn)
	profs, err := helper.GetAllDoc[[]model.Profile](config.Mongoconn, "profile")
	if err != nil {
		resp.Response = err.Error()
	} else {
		for _, prof := range profs {
			dt := &model.WebHook{
				URL:    prof.URL,
				Secret: prof.Secret,
			}
			res, err := helper.RefreshToken(dt, prof.Phonenumber, config.WAAPIGetToken, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
				break
			} else {
				resp.Response = helper.Jsonstr(res.ModifiedCount)
				httpstatus = http.StatusOK
			}
		}
	}
	//kirim report ke group
	filter := bson.M{"_id": report.TodayFilter()}
	wagroupidlist, err := atdb.GetAllDistinctDoc(config.Mongoconn, filter, "project.wagroupid", "pushrepo")
	if err != nil {
		resp.Info = "Gagal Query Distincs project.wagroupid"
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	for _, gid := range wagroupidlist {
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
			Messages: report.GetDataRepoMasukHariIni(config.Mongoconn, groupID) + "\n" + report.GetDataLaporanMasukHariini(config.Mongoconn, groupID),
		}
		resp, err = helper.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
		if err != nil {
			resp.Info = "Tidak berhak"
			resp.Response = err.Error()
			helper.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
	}
	helper.WriteJSON(respw, httpstatus, resp)
}

func NotFound(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	resp.Response = "Not Found"
	helper.WriteJSON(respw, http.StatusNotFound, resp)
}
