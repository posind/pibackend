package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/watoken"
)

func GetDataUser(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, helper.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = helper.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		helper.WriteJSON(respw, http.StatusNotFound, docuser)
		return
	}
	docuser.Name = payload.Alias
	helper.WriteJSON(respw, http.StatusOK, docuser)
}

func PostDataUser(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, helper.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = helper.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}
	var usr model.Userdomyikado
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		helper.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		usr.PhoneNumber = payload.Id
		usr.Name = payload.Alias
		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
		if err != nil {
			var respn model.Response
			respn.Status = "Gagal Insert Database"
			respn.Response = err.Error()
			helper.WriteJSON(respw, http.StatusNotModified, respn)
			return
		}
		usr.ID = idusr
		helper.WriteJSON(respw, http.StatusOK, usr)
		return
	}
	docuser.Name = payload.Alias
	docuser.Email = usr.Email
	docuser.GitHostUsername = usr.GitHostUsername
	docuser.GitlabUsername = usr.GitlabUsername
	docuser.GithubUsername = usr.GithubUsername
	atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id}, docuser)
	helper.WriteJSON(respw, http.StatusOK, docuser)
}
