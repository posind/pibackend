package lms

import (
	"strconv"

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
	newxs, newls, newbar, err := GetNewCookie(profile.Xsrf, profile.Lsession)
	if err != nil {
		return
	}
	newdt := &LoginProfile{
		Username: "madep",
		Bearer:   newbar,
		Xsrf:     newxs,
		Lsession: newls,
	}
	_, err = atdb.ReplaceOneDoc(db, "lmscreds", bson.M{"username": "madep"}, newdt)
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

	url := "https://pamongdesa.id/webservice/user?page=1&perpage=1&search=&role%5B%5D=2&role%5B%5D=3&role%5B%5D=4&role%5B%5D=5&role%5B%5D=6&sub_position=&verification=&approval=&province=&regency=&district=&village=&start_date=&end_date=&statuslogin=%0A"
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
	url := "https://pamongdesa.id/webservice/user?page=1&perpage=" + strconv.Itoa(total) + "&search=&role%5B%5D=2&role%5B%5D=3&role%5B%5D=4&role%5B%5D=5&role%5B%5D=6&sub_position=&verification=&approval=&province=&regency=&district=&village=&start_date=&end_date=&statuslogin=%0A"
	profile, err := atdb.GetOneDoc[LoginProfile](db, "lmscreds", bson.M{})
	if err != nil {
		return
	}
	_, res, err := atapi.GetWithBearer[Root](profile.Bearer, url)
	if err != nil {
		return
	}
	users = res.Data.Data
	return
}
