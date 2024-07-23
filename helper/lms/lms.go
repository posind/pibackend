package lms

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RefreshCookie(db *mongo.Database) (err error) {
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return
	}
	newxs, newls, newbar, err := GetNewCookie(profile.Xsrf, profile.Lsession, db)
	if err != nil {
		return
	}
	profile.Bearer = newbar
	profile.Xsrf = newxs
	profile.Lsession = newls
	_, err = atdb.ReplaceOneDoc(db, "lmscreds", bson.M{"username": "madep"}, profile)
	if err != nil {
		return
	}
	return

}

func GetTotalUser(db *mongo.Database) (total int, err error) {
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return
	}
	url := profile.URLUsers
	url = strings.ReplaceAll("1", "##TOTAL##", url)

	_, res, err := atapi.GetWithBearer[Root](profile.Bearer, url)
	if err != nil {
		return
	}
	total = res.Data.Meta.Total
	return
}

func GetAllUser(db *mongo.Database) (users []User, err error) {
	total, err := GetTotalUser(db)
	if err != nil {
		return
	}
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return
	}
	url := profile.URLUsers
	url = strings.ReplaceAll(strconv.Itoa(total), "##TOTAL##", url)
	_, res, err := atapi.GetWithBearer[Root](profile.Bearer, url)
	if err != nil {
		err = errors.New("GetWithBearer:" + err.Error() + profile.Bearer + " " + url)
		return
	}
	users = res.Data.Data
	return
}
