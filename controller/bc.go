package controller

import (
	"context"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDataSenders(respw http.ResponseWriter, req *http.Request) {
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
	_, err = atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	existingprjs, err := atdb.GetAllDoc[[]model.SenderDasboard](config.Mongoconn, "sender", primitive.M{})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotFound, respn)
		return
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = "Kakak belum input sender, silahkan input dulu ya"
		at.WriteJSON(respw, http.StatusNotFound, respn)
		return
	}
	at.WriteJSON(respw, http.StatusOK, existingprjs)
}

func GetDataSendersTerblokir(respw http.ResponseWriter, req *http.Request) {
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
	_, err = atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	existingprjs, err := atdb.GetAllDoc[[]model.SenderDasboard](config.Mongoconn, "bin", primitive.M{})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotFound, respn)
		return
	}
	if len(existingprjs) == 0 {
		var respn model.Response
		respn.Status = "Error : Data senders tidak di temukan"
		respn.Response = "Kakak belum input sender, silahkan input dulu ya"
		at.WriteJSON(respw, http.StatusNotFound, respn)
		return
	}
	at.WriteJSON(respw, http.StatusOK, existingprjs)
}

func GetRekapBlast(respw http.ResponseWriter, req *http.Request) {
	var respn model.Response
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}
	//check eksistensi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		docuser.PhoneNumber = payload.Id
		docuser.Name = payload.Alias
		at.WriteJSON(respw, http.StatusNotFound, docuser)
		return
	}
	docuser.Name = payload.Alias
	//melakukan pengambilan data belum terlayani
	// Menghitung jumlah dokumen dalam koleksi
	countqueue, err := config.Mongoconn.Collection("peserta").CountDocuments(context.TODO(), nil)
	if err != nil {
		respn.Status = "Error : penghitungan data queue"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusConflict, respn)
		return
	}
	countsent, err := config.Mongoconn.Collection("sent").CountDocuments(context.TODO(), nil)
	if err != nil {
		respn.Status = "Error : penghitungan data sent"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusConflict, respn)
		return
	}

	rekap := model.HelpdeskRekap{
		ToDo: int(countqueue),
		Done: int(countsent),
		All:  int(countqueue + countsent),
	}
	at.WriteJSON(respw, http.StatusOK, rekap)
}
