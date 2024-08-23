package helpdesk

import (
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/tiket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// phonefield: userphone or adminphone
func IsTicketClosed(phonefield string, phonenumber string, db *mongo.Database) (closed bool, tiket tiket.Bantuan, err error) {
	tiket, err = atdb.GetOneLatestDoc[tiket.Bantuan](db, "tiket", bson.M{"terlayani": bson.M{"$exists": false}, phonefield: phonenumber})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//tiket udah close karena no doc
			closed = true
			err = nil // Reset err ke nil karena ini bukan error, hanya kondisi normal
			return
		}
		// Jika ada error lain, kita return error tersebut
		return
	}
	// Jika ada tiket yang belum closed, kita kembalikan nilai default closed = false
	return
}

func InserNewTicket(userphone string, adminname string, adminphone string, db *mongo.Database) (err error) {
	dataapi := GetDataFromAPI(userphone)
	tiketbaru := tiket.Bantuan{
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
