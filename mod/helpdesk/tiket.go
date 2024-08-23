package helpdesk

import (
	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// phonefield: userphone or adminphone
func IsTicketClosed(phonefield string, phonenumber string, db *mongo.Database) (closed bool, tiket Bantuan, err error) {
	tiket, err = atdb.GetOneLatestDoc[Bantuan](db, "tiket", bson.M{"terlayani": bson.M{"$exists": false}, phonefield: phonenumber})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//tiket udah close karena no doc
			closed = true
			err = nil
		}
		return
	}
	// Jika ada tiket yang belum closed
	closed = false
	return
}

func InserNewTicket(userphone string, adminname string, adminphone string, db *mongo.Database) (err error) {
	dataapi := GetDataFromAPI(userphone)
	tiketbaru := Bantuan{
		UserName:   dataapi.Data.Fullname,
		UserPhone:  userphone,
		AdminPhone: adminphone,
		AdminName:  adminname,
		Prov:       dataapi.Data.Province,
		KabKot:     dataapi.Data.Regency,
		Kec:        dataapi.Data.District,
		Desa:       dataapi.Data.Village,
	}
	_, err = atdb.InsertOneDoc(db, "tiket", tiketbaru)
	if err != nil {
		return
	}
	return
}
