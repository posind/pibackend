package helpdesk

import (
	"context"
	"strconv"
	"time"

	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// check session udah ada atau belum kalo sudah ada maka refresh session
func CheckSession(phonenumber string, db *mongo.Database) (session Session, result bool, err error) {
	session, err = atdb.GetOneDoc[Session](db, "session", bson.M{"phonenumber": phonenumber})
	session.CreatedAt = time.Now()
	session.PhoneNumber = phonenumber
	if err != nil { //insert session klo belum ada
		_, err = db.Collection("session").InsertOne(context.TODO(), session)
		if err != nil {
			return
		}
	} else { //jika sesssion udah ada
		//refresh waktu session dengan waktu sekarang
		_, err = atdb.DeleteManyDocs(db, "session", bson.M{"phonenumber": phonenumber})
		if err != nil {
			return
		}
		_, err = db.Collection("session").InsertOne(context.TODO(), session)
		if err != nil {
			return
		}
		result = true
	}
	return
}

func GetMenuFromKeywordAndSetSession(keyword string, session Session, db *mongo.Database) (msg string, err error) {
	dt, err := atdb.GetOneDoc[Menu](db, "menu", bson.M{"keyword": keyword})
	if err != nil {
		return
	}
	atdb.UpdateOneDoc(db, "session", bson.M{"phonenumber": session.PhoneNumber}, bson.M{"list": dt.List})
	msg = dt.Header + "\n"
	for _, item := range dt.List {
		msg += strconv.Itoa(item.No) + ". " + item.Konten + "\n"
	}
	msg += dt.Footer
	return
}
