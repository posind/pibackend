package module

import (
	"context"
	"strconv"
	"time"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/mod"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func MenuSessionHandler(Profile itmodel.Profile, msg itmodel.IteungMessage, db *mongo.Database) string {
	//check apakah ada session, klo ga ada kasih reply menu
	Sesdoc, ses, err := CheckSession(msg.Phone_number, db)
	if err != nil {
		return err.Error()
	}
	if !ses { //jika tidak ada session atau session=false maka return menu dan update session isi list nomor menunya
		msg, err := GetMenuFromKeywordAndSetSession("menu", Sesdoc, db)
		if err != nil {
			return err.Error()
		}
		return msg

	}
	//jika ada session maka cek menu
	//check apakah pesan integer
	menuno, err := strconv.Atoi(msg.Message)
	if err == nil { //kalo nomor
		for _, menu := range Sesdoc.Menulist {
			if menuno == menu.No {
				msgs, err := GetMenuFromKeywordAndSetSession(menu.Keyword, Sesdoc, db)
				if err != nil {
					//jika di coll menu tidak ada maka akan proses keyword ke module
					msg.Message = msgs
					modname, _, _ := GetModuleName(Profile.Phonenumber, msg, db, "module")
					if modname == "" { //jika module tidak ada keyword tersebut maka langsung saja diarahkan ke faq
						dt, err := atdb.GetOneDoc[Datasets](db, "faq", bson.M{"question": menu.Keyword})
						if err != nil {
							return "modname: " + modname + "\npesan: " + msg.Message + "\nerror getmenufromkeywordandsession:" + err.Error()
						}
						return dt.Answer
					} else {
						return mod.Caller(Profile, modname, msg, db)

					}
				}
				return msgs
			}
		}
		return "Mohon maaf nomor menu yang anda masukkan tidak ada di daftar menu"
	}
	//kalo bukan nomor return empty
	return ""
}

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
