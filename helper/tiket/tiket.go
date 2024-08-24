package tiket

import (
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/waktu"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func UpdateUserMsgInTiket(userphone string, usermsg string, db *mongo.Database) (err error) {
	tiket, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"terlayani": bson.M{"$exists": false}, "userphone": userphone})
	if err != nil {
		return
	}
	wkt, err := waktu.GetDateTimeJKTNow()
	if err != nil {
		return
	}

	tiket.UserMessage += "\n" + wkt + " : " + usermsg
	_, err = atdb.ReplaceOneDoc(db, "tiket", bson.M{"_id": tiket.ID}, tiket)
	if err != nil {
		return
	}
	return
}

func GetNamaAdmin(adminphone string, db *mongo.Database) (name string) {
	tiket, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"adminphone": adminphone})
	if err != nil {
		return
	}
	return tiket.AdminName
}

func UpdateAdminMsgInTiket(adminphone string, adminmsg string, db *mongo.Database) (err error) {
	tiket, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"terlayani": bson.M{"$exists": false}, "adminphone": adminphone})
	if err != nil {
		return
	}
	wkt, err := waktu.GetDateTimeJKTNow()
	if err != nil {
		return
	}

	tiket.AdminMessage += "\n" + wkt + " : " + adminmsg
	_, err = atdb.ReplaceOneDoc(db, "tiket", bson.M{"_id": tiket.ID}, tiket)
	if err != nil {
		return
	}
	return
}

func IsAdmin(adminphone string, db *mongo.Database) (isadmin bool) {
	_, err := atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"adminphone": adminphone})
	if err != nil {
		return
	}
	isadmin = true
	return
}
