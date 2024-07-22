package controller

import (
	"net/http"
	"strconv"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/model"
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
