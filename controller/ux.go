package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/gcallapi"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	res, err := report.TambahPoinPresensibyPhoneNumber(config.Mongoconn, presensi.PhoneNumber, presensi.Lokasi, presensi.Skor, "presensi")
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

func PostRatingLaporan(respw http.ResponseWriter, req *http.Request) {
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
	hasil, err := atdb.GetOneLatestDoc[report.Laporan](config.Mongoconn, "uxlaporan", primitive.M{"_id": objectId})
	if err != nil {
		respn.Status = "Error : Data laporan tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	filter := bson.M{"_id": bson.M{"$eq": hasil.ID}}
	fields := bson.M{
		"rating":   rating.Rating,
		"komentar": rating.Komentar,
	}
	res, err := atdb.UpdateOneDoc(config.Mongoconn, "uxlaporan", filter, fields)
	if err != nil {
		respn.Status = "Error : Data laporan tidak berhasil di update data rating"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	poin := float64(rating.Rating) / 5.0
	_, err = report.TambahPoinLaporanbyPhoneNumber(config.Mongoconn, hasil.Project, hasil.NoPetugas, poin, "rating")
	if err != nil {
		respn.Info = "TambahPoinPushRepobyGithubUsername gagal"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusExpectationFailed, respn)
		return
	}
	message := "*" + hasil.Petugas + "*\nsudah dinilai oleh *" + hasil.Nama + " " + hasil.Phone + "* dengan rating *" + strconv.Itoa(rating.Rating) + "* komentar:\n" + rating.Komentar
	dt := &whatsauth.TextMessage{
		To:       hasil.NoPetugas,
		IsGroup:  false,
		Messages: message,
	}
	resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	respn.Response = strconv.Itoa(int(res.ModifiedCount))
	respn.Info = hasil.Nama
	at.WriteJSON(respw, http.StatusOK, respn)
}

func GetLaporan(respw http.ResponseWriter, req *http.Request) {
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
	hasil, err := atdb.GetOneLatestDoc[report.Laporan](config.Mongoconn, "uxlaporan", primitive.M{"_id": objectId})
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Data laporan tidak di temukan"
		respn.Response = err.Error()
		at.WriteJSON(respw, http.StatusNotImplemented, respn)
		return
	}
	at.WriteJSON(respw, http.StatusOK, hasil)
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
	var lap report.Laporan
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
		resp.Info = "TambahPoinPushRepobyGithubUsername gagal"
		resp.Response = err.Error()
		at.WriteJSON(w, http.StatusExpectationFailed, resp)
		return
	}
	message := "*Meeting " + event.Summary + "*\n" + lap.Kode + "\nAgenda: " + event.Description + "\nTanggal: " + event.Date + "\nJam: " + event.TimeStart + " - " + event.TimeEnd + "\nPembuat : " + docuser.Name + "\nURL Risalah Pertemuan: " + "https://www.do.my.id/rate/#" + lap.ID.Hex()
	dt := &whatsauth.TextMessage{
		To:       lap.Phone,
		IsGroup:  false,
		Messages: message,
	}
	resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
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
	var lap report.Laporan
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
	resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
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
	var lap report.Laporan
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
	resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
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
