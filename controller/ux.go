package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/gcallapi"
	"github.com/gocroot/helper/kimseok"
	"github.com/gocroot/helper/lms"
	"github.com/gocroot/helper/phone"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/tiket"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/mod/helpdesk"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostTaskList(w http.ResponseWriter, r *http.Request) {
	var resp itmodel.Response
	prof, err := whatsauth.GetAppProfile(at.GetParam(r), config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusBadRequest, resp)
		return
	}
	if at.GetSecretFromHeader(r) != prof.Secret {
		resp.Response = "Salah secret: " + at.GetSecretFromHeader(r)
		at.WriteJSON(w, http.StatusUnauthorized, resp)
		return
	}
	var tasklists []report.TaskList
	err = json.NewDecoder(r.Body).Decode(&tasklists)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusBadRequest, resp)
		return
	}
	docusr, err := atdb.GetOneLatestDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": tasklists[0].PhoneNumber})
	if err != nil {
		resp.Response = "Error : user tidak di temukan " + err.Error()
		at.WriteJSON(w, http.StatusForbidden, resp)
		return
	}
	lapuser, err := atdb.GetOneLatestDoc[model.Laporan](config.Mongoconn, "uxlaporan", primitive.M{"_id": tasklists[0].LaporanID})
	if err != nil {
		resp.Response = "Error : user tidak di temukan " + err.Error()
		at.WriteJSON(w, http.StatusForbidden, resp)
		return
	}
	for _, task := range tasklists {
		task.ProjectID = lapuser.Project.ID
		task.ProjectName = lapuser.Project.Name
		task.Email = docusr.Email
		task.UserID = docusr.ID
		task.MeetID = lapuser.MeetID
		task.MeetGoal = lapuser.MeetEvent.Summary
		task.MeetDate = lapuser.MeetEvent.Date
		task.ProjectWAGroupID = lapuser.Project.WAGroupID
		_, err = atdb.InsertOneDoc(config.Mongoconn, "tasklist", task)
		if err != nil {
			resp.Info = "Kakak sudah melaporkan tasklist sebelumnya"
			resp.Response = "Error : tidak bisa insert ke database " + err.Error()
			at.WriteJSON(w, http.StatusForbidden, resp)
			return
		}
	}
	res, err := report.TambahPoinTasklistbyPhoneNumber(config.Mongoconn, docusr.PhoneNumber, lapuser.Project, float64(len(tasklists)), "tasklist")
	if err != nil {
		resp.Info = "Tambah Poin Tasklist gagal"
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusExpectationFailed, resp)
		return
	}
	resp.Response = strconv.Itoa(int(res.ModifiedCount))
	resp.Info = docusr.Name
	at.WriteJSON(w, http.StatusOK, resp)
}

func PostPresensi(respw http.ResponseWriter, req *http.Request) {
	var resp itmodel.Response
	prof, err := whatsauth.GetAppProfile(at.GetParam(req), config.Mongoconn)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	if at.GetSecretFromHeader(req) != prof.Secret {
		resp.Response = "Salah secret: " + at.GetSecretFromHeader(req)
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	var presensi report.PresensiDomyikado
	err = json.NewDecoder(req.Body).Decode(&presensi)
	if err != nil {
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	docusr, err := atdb.GetOneLatestDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": presensi.PhoneNumber})
	if err != nil {
		resp.Response = "Error : user tidak di temukan " + err.Error()
		at.WriteJSON(respw, http.StatusForbidden, resp)
		return
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "presensi", presensi)
	if err != nil {
		resp.Info = "Kakak sudah melaporkan presensi sebelumnya"
		resp.Response = "Error : tidak bisa insert ke database " + err.Error()
		at.WriteJSON(respw, http.StatusForbidden, resp)
		return
	}
	res, err := report.TambahPoinPresensibyPhoneNumber(config.Mongoconn, presensi.PhoneNumber, presensi.Lokasi, presensi.Skor, config.WAAPIToken, config.WAAPIMessage, "presensi")
	if err != nil {
		resp.Info = "Tambah Poin Presensi gagal"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusExpectationFailed, resp)
		return
	}
	resp.Response = strconv.Itoa(int(res.ModifiedCount))
	resp.Info = docusr.Name
	at.WriteJSON(respw, http.StatusOK, resp)
}

// testimoni dari useng lms pamong
func PostTestimoni(respw http.ResponseWriter, req *http.Request) {
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(req))
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Token Tidak Valid "
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Decode Token Error: " + at.GetLoginFromHeader(req)
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusForbidden, respn)
		return
	}
	userdt := lms.GetDataFromAPI(payload.Id)
	if userdt.Data.Fullname == "" {
		at.WriteJSON(respw, http.StatusNotFound, userdt)
		return
	}
	//pindah ke struck user
	var usersub model.Peserta
	usersub.Fullname = userdt.Data.Fullname
	usersub.Desa = userdt.Data.Village
	usersub.Kec = userdt.Data.District
	usersub.Kab = userdt.Data.Regency
	usersub.PhoneNumber = payload.Id
	usersub.Provinsi = userdt.Data.Province

	var rating report.Rating
	var respn model.Response
	err = json.NewDecoder(req.Body).Decode(&rating)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	usersub.Rating = rating.Rating
	usersub.Komentar = rating.Komentar
	res, err := atdb.InsertOneDoc(config.Mongoconn, "unsubs", usersub)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	respn.Response = res.Hex()
	respn.Info = usersub.Fullname
	at.WriteJSON(respw, http.StatusOK, respn)
}

// mendapatkan random testi 4 buah untuk halaman depan
func GetRandomTesti4(respw http.ResponseWriter, req *http.Request) {
	var respn model.Response
	lstpeserta, err := atdb.GetRandomDoc[model.Peserta](config.Mongoconn, "unsubs", 4)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	var listtesti []model.Testi
	for _, testi := range lstpeserta {
		tst := model.Testi{
			Isi:    testi.Komentar,
			Nama:   testi.Fullname,
			Daerah: "Desa " + testi.Desa + " Kec. " + testi.Kec + " " + testi.Kab + " Prov. " + testi.Provinsi,
		}
		listtesti = append(listtesti, tst)
	}
	testidepan := model.Depan{
		List: listtesti,
	}
	at.WriteJSON(respw, http.StatusOK, testidepan)
}

// feedback dan meeting jadi satu disini
func PostUnsubscribe(respw http.ResponseWriter, req *http.Request) {
	var rating report.Rating
	var respn model.Response
	err := json.NewDecoder(req.Body).Decode(&rating)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	objectId, err := primitive.ObjectIDFromHex(rating.ID)
	if err != nil {
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	hasil, err := atdb.GetOneLatestDoc[model.Peserta](config.Mongoconn, "sent", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data laporan tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	hasil.Rating = rating.Rating
	hasil.Komentar = rating.Komentar
	res, err := atdb.InsertOneDoc(config.Mongoconn, "unsubs", hasil)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	respn.Response = res.Hex()
	respn.Info = hasil.Fullname
	at.WriteJSON(respw, http.StatusOK, respn)
}

// mendapatkan data FAQ
func GetFAQ(respw http.ResponseWriter, req *http.Request) {
    var filter primitive.M = primitive.M{}

    // Ambil query parameters
    query := req.URL.Query()
    id := query.Get("id")
    question := query.Get("question")
    limitParam := query.Get("limit")

    // Filter berdasarkan ID jika disediakan
    if id != "" {
        objectId, err := primitive.ObjectIDFromHex(id)
        if err != nil {
            var respn model.Response
            respn.Status = "Error: ObjectID Tidak Valid"
            respn.Info = at.GetSecretFromHeader(req)
            respn.Location = "Encode Object ID Error"
            respn.Response = err.Error()
            at.WriteJSON(respw, http.StatusBadRequest, respn)
            return
        }
        filter["_id"] = objectId
    }

    // Filter berdasarkan pertanyaan jika disediakan
    if question != "" {
        filter["question"] = primitive.Regex{Pattern: question, Options: "i"} // Regex untuk pencarian case-insensitive
    }

    // Parsing limit jika diberikan
    var limit int64 = 20
    if limitParam != "" {
        if parsedLimit, err := strconv.ParseInt(limitParam, 10, 64); err == nil {
            limit = parsedLimit
        }
    }

    // Ambil data dari database dengan filter dan limit
    findOptions := options.Find().SetLimit(limit)
    var items []kimseok.Datasets
    collection := config.Mongoconn.Collection("faq")
    cursor, err := collection.Find(context.Background(), filter, findOptions)
    if err != nil {
        var respn model.Response
        respn.Status = "Error: Internal Server Error"
        respn.Info = "Terjadi kesalahan saat mengambil data dari database."
        at.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }
    defer cursor.Close(context.Background())

    // Decode hasil pencarian
    if err = cursor.All(context.Background(), &items); err != nil {
        var respn model.Response
        respn.Status = "Error: Internal Server Error"
        respn.Info = "Terjadi kesalahan saat memproses data."
        at.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    // Periksa apakah ada hasil
    if len(items) == 0 {
        var respn model.Response
        respn.Status = "Error: Tidak Ada Data"
        respn.Info = "Tidak ada data FAQ yang sesuai dengan filter."
        at.WriteJSON(respw, http.StatusNotFound, respn)
        return
    }

    // Filter hanya question dan answer untuk dikirim ke frontend
    type FAQResponse struct {
        Question string `json:"question"`
        Answer   string `json:"answer"`
    }

    var faqResponses []FAQResponse
    for _, item := range items {
        faqResponses = append(faqResponses, FAQResponse{
            Question: item.Question,
            Answer:   item.Answer,
        })
    }

    // Kirim hasil ke client
    at.WriteJSON(respw, http.StatusOK, faqResponses)
}

// Tambah FAQ
func AddFAQ(respw http.ResponseWriter, req *http.Request) {
    var newFAQ kimseok.Datasets

    // Decode body request
    if err := json.NewDecoder(req.Body).Decode(&newFAQ); err != nil {
        var respn model.Response
        respn.Status = "Error: Invalid JSON"
        respn.Response = err.Error()
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Validasi field question dan answer
    if newFAQ.Question == "" || newFAQ.Answer == "" {
        var respn model.Response
        respn.Status = "Error: Field Tidak Lengkap"
        respn.Response = "Field Question dan Answer harus diisi."
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Masukkan data ke koleksi "faq"
    collection := config.Mongoconn.Collection("faq")
    result, err := collection.InsertOne(context.TODO(), newFAQ)
    if err != nil {
        var respn model.Response
        respn.Status = "Error: Gagal Menambahkan Data"
        respn.Response = err.Error()
        at.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    // Respons sukses
    var respn model.Response
    respn.Status = "Sukses: FAQ Ditambahkan"
    respn.Response = fmt.Sprintf("Data berhasil ditambahkan dengan ID: %v", result.InsertedID)
    at.WriteJSON(respw, http.StatusOK, respn)
}




// update FAQ
// Update FAQ
func UpdateFAQ(respw http.ResponseWriter, req *http.Request) {
    // Ambil query parameters
    query := req.URL.Query()
    id := query.Get("id")

    if id == "" {
        var respn model.Response
        respn.Status = "Error: ID Tidak Diberikan"
        respn.Response = "Parameter ID diperlukan untuk update."
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    objectId, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        var respn model.Response
        respn.Status = "Error: ObjectID Tidak Valid"
        respn.Response = err.Error()
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Decode body request
    var updateData struct {
        Question string `json:"question"`
        Answer   string `json:"answer"`
    }

    if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
        var respn model.Response
        respn.Status = "Error: Invalid JSON"
        respn.Response = err.Error()
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Validasi field question dan answer
    if updateData.Question == "" || updateData.Answer == "" {
        var respn model.Response
        respn.Status = "Error: Field Tidak Lengkap"
        respn.Response = "Field Question dan Answer harus diisi."
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Update dokumen di database
    filter := primitive.M{"_id": objectId}
    update := primitive.M{
        "$set": primitive.M{
            "question": updateData.Question,
            "answer":   updateData.Answer,
        },
    }

    result, err := config.Mongoconn.Collection("faq").UpdateOne(context.TODO(), filter, update)
    if err != nil {
        var respn model.Response
        respn.Status = "Error: Gagal Update Data"
        respn.Response = err.Error()
        at.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    if result.ModifiedCount == 0 {
        var respn model.Response
        respn.Status = "Error: Tidak Ada Data Yang Diperbarui"
        respn.Response = "Pastikan filter valid dan data berbeda dari sebelumnya."
        at.WriteJSON(respw, http.StatusNotFound, respn)
        return
    }

    // Respons sukses
    var respn model.Response
    respn.Status = "Sukses: FAQ Diperbarui"
    respn.Response = fmt.Sprintf("Data berhasil diperbarui. ID: %v", id)
    at.WriteJSON(respw, http.StatusOK, respn)
}

// Hapus FAQ
func DeleteFAQ(respw http.ResponseWriter, req *http.Request) {
    // Ambil query parameters
    query := req.URL.Query()
    id := query.Get("id")

    if id == "" {
        var respn model.Response
        respn.Status = "Error: ID Tidak Diberikan"
        respn.Response = "Parameter ID diperlukan untuk menghapus data."
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    objectId, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        var respn model.Response
        respn.Status = "Error: ObjectID Tidak Valid"
        respn.Response = err.Error()
        at.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Hapus dokumen di database
    filter := primitive.M{"_id": objectId}
    result, err := config.Mongoconn.Collection("faq").DeleteOne(context.TODO(), filter)
    if err != nil {
        var respn model.Response
        respn.Status = "Error: Gagal Menghapus Data"
        respn.Response = err.Error()
        at.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    if result.DeletedCount == 0 {
        var respn model.Response
        respn.Status = "Error: Tidak Ada Data Yang Dihapus"
        respn.Response = "Data dengan ID tersebut tidak ditemukan."
        at.WriteJSON(respw, http.StatusNotFound, respn)
        return
    }

    // Respons sukses
    var respn model.Response
    respn.Status = "Sukses: FAQ Dihapus"
    respn.Response = fmt.Sprintf("Data berhasil dihapus. ID: %v", id)
    at.WriteJSON(respw, http.StatusOK, respn)
}



// mendapatkan user yang sent dan mau unnsubscribe
func GetSentItem(respw http.ResponseWriter, req *http.Request) {
	id := at.GetParam(req)
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	hasil, err := atdb.GetOneLatestDoc[model.Peserta](config.Mongoconn, "sent", primitive.M{"_id": objectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data profile user sent tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	hasil.PhoneNumber = ""
	at.WriteJSON(respw, http.StatusOK, hasil)
}

// mendapatkan tiket yang sudah closed Profile, err := atdb.GetOneDoc[itmodel.Profile](Mongoconn, "profile", primitive.M{"phonenumber": PhoneNumber})
func GetClosedTicket(respw http.ResponseWriter, req *http.Request) {
	id := at.GetParam(req)
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	hasil, err := atdb.GetOneLatestDoc[tiket.Bantuan](config.Mongoconn, "tiket", primitive.M{"_id": objectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data tiket tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}

	at.WriteJSON(respw, http.StatusOK, hasil)
}

// mendapatkan semua list bot yang aktif
func GetBotList(respw http.ResponseWriter, req *http.Request) {
	Profiles, err := atdb.GetAllDoc[[]itmodel.Profile](config.Mongoconn, "profile", primitive.M{})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data tiket tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	var phonelist []string

	for _, profile := range Profiles {
		phonelist = append(phonelist, phone.MaskPhoneNumber(profile.Phonenumber))
	}
	hasil := model.PhoneList{PhoneList: phonelist}

	at.WriteJSON(respw, http.StatusOK, hasil)
}

// feedback dari tiket yang sudah tertutup
func PostMasukanTiket(respw http.ResponseWriter, req *http.Request) {
	var rating report.Rating
	var respn model.Response
	err := json.NewDecoder(req.Body).Decode(&rating)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	objectId, err := primitive.ObjectIDFromHex(rating.ID)
	if err != nil {
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = at.GetSecretFromHeader(req)
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	updatefields := primitive.M{
		"ratelayanan": rating.Rating,
		"masukan":     rating.Komentar,
	}
	res, err := atdb.UpdateOneDoc(config.Mongoconn, "tiket", primitive.M{"_id": objectId}, updatefields)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	//mendapatkan document untuk informasi ke admin
	hasil, err := atdb.GetOneLatestDoc[tiket.Bantuan](config.Mongoconn, "tiket", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data laporan tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}

	nama := hasil.UserName
	if nama == "" {
		nama = phone.MaskPhoneNumber(hasil.UserPhone)
	}

	respn.Response = strconv.Itoa(int(res.ModifiedCount))
	respn.Info = nama
	//info ke admin
	message := helpdesk.GetPrefillMessage("adminnotiffeedback", config.Mongoconn)
	message = fmt.Sprintf(message, rating.Rating, nama, rating.Komentar)
	dt := &whatsauth.TextMessage{
		To:       hasil.AdminPhone,
		IsGroup:  false,
		Messages: message,
	}
	go atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)

	at.WriteJSON(respw, http.StatusOK, respn)
}

func PostMeeting(w http.ResponseWriter, r *http.Request) {
	var respn model.Response
	//otorisasi dan validasi inputan
	payload, err := watoken.Decode(config.PublicKeyWhatsAuth, at.GetLoginFromHeader(r))
	if err != nil {
		respn.Status = "Error : Token Tidak Valid"
		respn.Info = at.GetSecretFromHeader(r)
		respn.Location = "Decode Token Error"
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}
	var event gcallapi.SimpleEvent
	err = json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusBadRequest, respn)
		return
	}
	//check validasi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		respn.Status = "Error : Data user tidak di temukan: " + payload.Id
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusNotImplemented, respn)
		return
	}
	prjuser, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": event.ProjectID})
	if err != nil {
		respn.Status = "Error : Data project tidak di temukan: " + event.ProjectID.Hex()
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusNotImplemented, respn)
		return
	}
	//lojik inputan post
	var lap model.Laporan
	lap.User = docuser
	lap.Project = prjuser
	lap.Phone = prjuser.Owner.PhoneNumber
	lap.Nama = prjuser.Owner.Name
	lap.Petugas = docuser.Name
	lap.NoPetugas = docuser.PhoneNumber
	lap.Solusi = event.Description
	//mengambil daftar email dari project member
	var attendees []string
	for _, member := range prjuser.Members {
		attendees = append(attendees, member.Email)
	}
	event.Attendees = attendees

	gevt, err := gcallapi.HandlerCalendar(config.Mongoconn, event)
	if err != nil {
		respn.Status = "Gagal Membuat Google Calendar"
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusNotModified, respn)
		return
	}
	_, err = atdb.InsertOneDoc(config.Mongoconn, "meetinglog", gevt)
	if err != nil {
		respn.Status = "Gagal Insert Database meetinglog"
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusNotModified, respn)
		return
	}
	event.ID, err = atdb.InsertOneDoc(config.Mongoconn, "meeting", event)
	if err != nil {
		respn.Status = "Gagal Insert Database meeting"
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusNotModified, respn)
		return
	}
	lap.MeetID = event.ID
	lap.MeetEvent = event
	lap.Kode = gevt.HtmlLink
	lap.ID, err = atdb.InsertOneDoc(config.Mongoconn, "uxlaporan", lap)
	if err != nil {
		respn.Status = "Gagal Insert Database"
		respn.Response = err.Error()
		at.WriteJSON(w, http.StatusNotModified, respn)
		return
	}
	_, err = report.TambahPoinLaporanbyPhoneNumber(config.Mongoconn, prjuser, docuser.PhoneNumber, 1, "meeting")
	if err != nil {
		var resp model.Response
		resp.Info = "TambahPoinLaporanbyPhoneNumber gagal"
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusExpectationFailed, resp)
		return
	}

	message := "*" + strings.TrimSpace(event.Summary) + "*\n" + lap.Kode + "\nLokasi:\n" + event.Location + "\nAgenda:\n" + event.Description + "\nTanggal: " + event.Date + "\nJam: " + event.TimeStart + " - " + event.TimeEnd + "\nNotulen : " + docuser.Name + "\nURL Input Risalah Pertemuan:\n" + "https://www.do.my.id/resume/#" + lap.ID.Hex()
	dt := &whatsauth.TextMessage{
		To:       lap.Project.WAGroupID,
		IsGroup:  true,
		Messages: message,
	}
	_, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusUnauthorized, resp)
		return
	}
	at.WriteJSON(w, http.StatusOK, lap)
}

func PostLaporan(respw http.ResponseWriter, req *http.Request) {
	//otorisasi dan validasi inputan
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
	var lap model.Laporan
	err = json.NewDecoder(req.Body).Decode(&lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	if lap.Solusi == "" {
		var respn model.Response
		respn.Status = "Error : Telepon atau nama atau solusi tidak diisi"
		respn.Response = "Isi lebih lengkap dahulu"
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	//check validasi user
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan: " + payload.Id
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	//ambil data project
	prjobjectId, err := primitive.ObjectIDFromHex(lap.Kode)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = lap.Kode
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	prjuser, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prjobjectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan: " + lap.Kode
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	//lojik inputan post
	lap.User = docuser
	lap.Project = prjuser
	lap.Phone = prjuser.Owner.PhoneNumber
	lap.Nama = prjuser.Owner.Name
	lap.Petugas = docuser.Name
	lap.NoPetugas = docuser.PhoneNumber

	idlap, err := atdb.InsertOneDoc(config.Mongoconn, "uxlaporan", lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal Insert Database"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotModified, respn)
		return
	}
	_, err = report.TambahPoinLaporanbyPhoneNumber(config.Mongoconn, prjuser, docuser.PhoneNumber, 1, "laporan")
	if err != nil {
		var resp model.Response
		resp.Info = "TambahPoinPushRepobyGithubUsername gagal"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusExpectationFailed, resp)
		return
	}
	message := "*Permintaan Feedback Pekerjaan*\n" + "Petugas : " + docuser.Name + "\nDeskripsi:" + lap.Solusi + "\n Beri Nilai: " + "https://www.do.my.id/rate/#" + idlap.Hex()
	dt := &whatsauth.TextMessage{
		To:       lap.Phone,
		IsGroup:  false,
		Messages: message,
	}
	_, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	at.WriteJSON(respw, http.StatusOK, lap)

}

func PostFeedback(respw http.ResponseWriter, req *http.Request) {
	//otorisasi dan validasi inputan
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
	var lap model.Laporan
	err = json.NewDecoder(req.Body).Decode(&lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	if lap.Phone == "" || lap.Nama == "" || lap.Solusi == "" {
		var respn model.Response
		respn.Status = "Error : Telepon atau nama atau solusi tidak diisi"
		respn.Response = "Isi lebih lengkap dahulu"
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	//validasi eksistensi user di db
	docuser, err := atdb.GetOneDoc[model.Userdomyikado](config.Mongoconn, "user", primitive.M{"phonenumber": payload.Id})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data user tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	//ambil data project
	prjobjectId, err := primitive.ObjectIDFromHex(lap.Kode)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : ObjectID Tidak Valid"
		respn.Info = lap.Kode
		respn.Location = "Encode Object ID Error"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusBadRequest, respn)
		return
	}
	prjuser, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"_id": prjobjectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data project tidak di temukan: " + lap.Kode
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	//lojik inputan post
	lap.Project = prjuser
	lap.User = docuser
	lap.Phone = ValidasiNoHP(lap.Phone)
	lap.Petugas = docuser.Name
	lap.NoPetugas = docuser.PhoneNumber

	idlap, err := atdb.InsertOneDoc(config.Mongoconn, "uxlaporan", lap)
	if err != nil {
		var respn model.Response
		respn.Status = "Gagal Insert Database"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotModified, respn)
		return
	}
	_, err = report.TambahPoinLaporanbyPhoneNumber(config.Mongoconn, prjuser, docuser.PhoneNumber, 1, "feedback")
	if err != nil {
		var resp model.Response
		resp.Info = "TambahPoinLaporanbyPhoneNumber gagal"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusExpectationFailed, resp)
		return
	}
	message := "*Permintaan Feedback*\n" + "Petugas : " + docuser.Name + "\nDeskripsi:" + lap.Solusi + "\n Beri Nilai: " + "https://www.do.my.id/rate/#" + idlap.Hex()
	dt := &whatsauth.TextMessage{
		To:       lap.Phone,
		IsGroup:  false,
		Messages: message,
	}
	_, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	at.WriteJSON(respw, http.StatusOK, lap)

}

func ValidasiNoHP(nomor string) string {
	nomor = strings.ReplaceAll(nomor, " ", "")
	nomor = strings.ReplaceAll(nomor, "+", "")
	nomor = strings.ReplaceAll(nomor, "-", "")
	return nomor
}
