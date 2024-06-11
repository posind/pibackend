package controller

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
)

func GetUserData(w http.ResponseWriter, r *http.Request) {
	var response itmodel.Response
	users, err := atdb.GetAllDoc[[]model.Userdomyikado](config.Mongoconn, "user", bson.M{})
	if(err != nil){
		response.Response = err.Error()
		at.WriteJSON(w, http.StatusBadRequest, response)
	}
	at.WriteJSON(w, http.StatusOK, users)
}