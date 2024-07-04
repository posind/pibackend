package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/helper/whatsauth"
)

func GetDataUser(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		at.WriteJSON(respw, http.StatusNotFound, docuser)
		return
	}
	docuser.Name = payload.Alias
	at.WriteJSON(respw, http.StatusOK, docuser)
}

func PostDataUser(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}
	var usr model.Userdomyikado
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
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
			at.WriteJSON(respw, http.StatusNotModified, respn)
			return
		}
		usr.ID = idusr
		at.WriteJSON(respw, http.StatusOK, usr)
		return
	}
	docuser.Name = payload.Alias
	docuser.Email = usr.Email
	docuser.GitHostUsername = usr.GitHostUsername
	docuser.GitlabUsername = usr.GitlabUsername
	docuser.GithubUsername = usr.GithubUsername
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id}, docuser)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal replaceonedoc"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusConflict, respn)
		return
	}
	//melakukan update di seluruh member project
	//ambil project yang member sebagai anggota
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"members._id": docuser.ID})
	if err != nil { //kalo belum jadi anggota project manapun aman langsung ok
		at.WriteJSON(respw, http.StatusOK, docuser)
		return
	}
	if len(existingprjs) == 0 { //kalo belum jadi anggota project manapun aman langsung ok
		at.WriteJSON(respw, http.StatusOK, docuser)
		return
	}
	//loop keanggotaan setiap project dan menggantinya dengan doc yang terupdate
	for _, prj := range existingprjs {
		memberToDelete := model.Userdomyikado{PhoneNumber: docuser.PhoneNumber}
		_, err := atdb.DeleteDocFromArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
		if err != nil {
			var respn model.Response
			respn.Status = "Error : Data project tidak di temukan"
			respn.Response = err.Error()
			at.WriteJSON(respw, http.StatusNotFound, respn)
			return
		}
		_, err = atdb.AddDocToArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", docuser)
		if err != nil {
			var respn model.Response
			respn.Status = "Error : Gagal menambahkan member ke project"
			respn.Response = err.Error()
			at.WriteJSON(respw, http.StatusExpectationFailed, respn)
			return
		}

	}

	at.WriteJSON(respw, http.StatusOK, docuser)
}

func PostDataUserFromWA(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	prof, err := whatsauth.GetAppProfile(at.GetParam(req), config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	if at.GetSecretFromHeader(req) != prof.Secret {
		resp.Response = "Salah secret: " + at.GetSecretFromHeader(req)
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	var usr model.Userdomyikado
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		resp.Response = "Error : Body tidak valid"
		resp.Info = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": usr.PhoneNumber})
	if err != nil {
		idusr, err := atdb.InsertOneDoc(config.Mongoconn, "user", usr)
		if err != nil {
			resp.Response = "Gagal Insert Database"
			resp.Info = err.Error()
			at.WriteJSON(respw, http.StatusNotModified, resp)
			return
		}
		resp.Info = idusr.Hex()
		at.WriteJSON(respw, http.StatusOK, resp)
		return
	}
	docuser.Name = usr.Name
	docuser.Email = usr.Email
	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "user", primitive.M{"phonenumber": usr.PhoneNumber}, docuser)
	if err != nil {
		resp.Response = "Gagal replaceonedoc"
		resp.Info = err.Error()
		at.WriteJSON(respw, http.StatusConflict, resp)
		return
	}
	//melakukan update di seluruh member project
	//ambil project yang member sebagai anggota
	existingprjs, err := atdb.GetAllDoc[[]model.Project](config.Mongoconn, "project", primitive.M{"members._id": docuser.ID})
	if err != nil { //kalo belum jadi anggota project manapun aman langsung ok
		resp.Response = "belum terdaftar di project manapun"
		at.WriteJSON(respw, http.StatusOK, resp)
		return
	}
	if len(existingprjs) == 0 { //kalo belum jadi anggota project manapun aman langsung ok
		resp.Response = "belum terdaftar di project manapun"
		at.WriteJSON(respw, http.StatusOK, resp)
		return
	}
	//loop keanggotaan setiap project dan menggantinya dengan doc yang terupdate
	for _, prj := range existingprjs {
		memberToDelete := model.Userdomyikado{PhoneNumber: docuser.PhoneNumber}
		_, err := atdb.DeleteDocFromArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", memberToDelete)
		if err != nil {
			resp.Response = "Error : Data project tidak di temukan"
			resp.Info = err.Error()
			at.WriteJSON(respw, http.StatusNotFound, resp)
			return
		}
		_, err = atdb.AddDocToArray[model.Userdomyikado](config.Mongoconn, "project", prj.ID, "members", docuser)
		if err != nil {
			resp.Response = "Error : Gagal menambahkan member ke project"
			resp.Info = err.Error()
			at.WriteJSON(respw, http.StatusExpectationFailed, resp)
			return
		}

	}
	resp.Info = docuser.ID.Hex()
	resp.Info = docuser.Email
	at.WriteJSON(respw, http.StatusOK, resp)
}
