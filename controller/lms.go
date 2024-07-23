package controller

import (
	"net/http"
	"strconv"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func GetLMSUser(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	total, err := lms.GetTotalUser(config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	resp.Info = strconv.Itoa(total)
	at.WriteJSON(respw, http.StatusOK, resp)
}

func CopyLMSUser(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	users, err := lms.GetAllUser(config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	_, err = atdb.InsertManyDocs[lms.User](config.Mongoconn, "lmsusers", users)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusConflict, resp)
		return
	}
	resp.Info = "ok"
	at.WriteJSON(respw, http.StatusOK, resp)
}

func GetCountDocUser(w http.ResponseWriter, r *http.Request) {
	var resp model.Response
	filter := bson.M{
		"profileapproved": 1,
		"roles":           "User",
	}
	count1, err := atdb.GetCountDoc(config.Mongoconn, "lmsusers", filter)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusConflict, resp)
		return
	}
	filter = bson.M{
		"profileapproved": 2,
		"roles":           "User",
	}
	count2, err := atdb.GetCountDoc(config.Mongoconn, "lmsusers", filter)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusConflict, resp)
		return
	}
	filter = bson.M{
		"profileapproved": 3,
		"roles":           "User",
	}
	count3, err := atdb.GetCountDoc(config.Mongoconn, "lmsusers", filter)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusConflict, resp)
		return
	}
	filter = bson.M{
		"profileapproved": 4,
		"roles":           "User",
	}
	count4, err := atdb.GetCountDoc(config.Mongoconn, "lmsusers", filter)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusConflict, resp)
		return
	}
	count5, err := atdb.GetCountDoc(config.Mongoconn, "lmsusers", bson.M{"roles": "User"})
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusConflict, resp)
		return
	}
	// 1. Belum Lengkap
	// 2. Menunggu Persetujuan
	// 3. Disetujui
	// 4. Ditolak
	rkp := lms.RekapitulasiUser{
		BelumLengkap:        count1,
		MenungguPersetujuan: count2,
		Disetujui:           count3,
		Ditolak:             count4,
		Total:               count5,
	}
	at.WriteJSON(w, http.StatusOK, rkp)

}

func DropLMSUser(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	err := atdb.DropCollection(config.Mongoconn, "lmsusers")
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusConflict, resp)
		return
	}
	resp.Info = "ok"
	at.WriteJSON(respw, http.StatusOK, resp)
}

func RefreshLMSCookie(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	err := lms.RefreshCookie(config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	resp.Info = "ok"
	at.WriteJSON(respw, http.StatusOK, resp)
}
