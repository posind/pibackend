package helpdesk

import (
	"context"
	"time"

	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// check session hub buat baru atau refresh session
func CheckHubSession(userphone, adminphone string, db *mongo.Database) (session SessionHub, result bool, err error) {
	session, err = atdb.GetOneDoc[SessionHub](db, "hub", bson.M{"userphone": userphone})
	session.CreatedAt = time.Now()
	if err != nil { //insert session klo belum ada
		session.UserPhone = userphone
		session.AdminPhone = adminphone
		_, err = db.Collection("hub").InsertOne(context.TODO(), session)
		if err != nil {
			return
		}
	} else { //jika sesssion udah ada
		//refresh waktu session dengan waktu sekarang
		_, err = atdb.DeleteManyDocs(db, "hub", bson.M{"userphone": userphone})
		if err != nil {
			return
		}
		_, err = db.Collection("hub").InsertOne(context.TODO(), session)
		if err != nil {
			return
		}
		result = true
	}
	return
}
