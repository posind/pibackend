package controller

import (
	"net/http"
	"strconv"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
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
