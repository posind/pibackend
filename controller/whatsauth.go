package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func GetHome(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	resp.Response = at.GetIPaddress()
	at.WriteJSON(respw, http.StatusOK, resp)
}

func PostInboxNomor(respw http.ResponseWriter, req *http.Request) {
	var resp whatsauth.Response
	var msg whatsauth.IteungMessage
	httpstatus := http.StatusUnauthorized
	resp.Response = "Wrong Secret"
	waphonenumber := at.GetParam(req)
	prof, err := whatsauth.GetAppProfile(waphonenumber, config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		httpstatus = http.StatusServiceUnavailable
	}
	if at.GetSecretFromHeader(req) == prof.Secret {
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			resp.Response = err.Error()
		} else {
			resp, err = whatsauth.WebHook(prof.QRKeyword, waphonenumber, config.WAAPIQRLogin, config.WAAPIMessage, msg, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
			}
		}
	}
	at.WriteJSON(respw, httpstatus, resp)
}

func GetNewToken(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	httpstatus := http.StatusServiceUnavailable
	//prof, err := helper.GetAppProfile(config.WAPhoneNumber, config.Mongoconn)
	profs, err := atdb.GetAllDoc[[]model.Profile](config.Mongoconn, "profile", bson.M{})
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, httpstatus, resp)
		return
	} else {
		for _, prof := range profs {
			dt := &whatsauth.WebHookInfo{
				URL:    prof.URL,
				Secret: prof.Secret,
			}
			res, err := whatsauth.RefreshToken(dt, prof.Phonenumber, config.WAAPIGetToken, config.Mongoconn)
			if err != nil {
				resp.Response = err.Error()
				break
			} else {
				resp.Response = at.Jsonstr(res.ModifiedCount)
				httpstatus = http.StatusOK
			}
		}
		//helper.WriteJSON(respw, httpstatus, resp)
		//return
	}
	//kirim report ke group
	report.RekapPagiHari(respw, req)

}

func NotFound(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	resp.Response = "Not Found"
	at.WriteJSON(respw, http.StatusNotFound, resp)
}
