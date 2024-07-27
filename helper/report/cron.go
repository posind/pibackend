package report

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"github.com/raykov/gofpdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RekapMeetingKemarin(db *mongo.Database, projectName string) (base64Str string, err error) {
	filter := CreateFilterMeetingYesterday(projectName, true)
	laporanDocs, err := atdb.GetAllDoc[[]Laporan](db, "uxlaporan", filter) //CreateFilterMeetingYesterday(projectName)
	if err != nil {
		return
	}
	if len(laporanDocs) == 0 {
		return
	}
	// Buat PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)

	// Menambahkan fungsi footer
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15) // Posisi footer dari bawah halaman
		pdf.SetFont("Arial", "I", 8)
		pageNo := pdf.PageNo() // Mendapatkan nomor halaman
		pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d", pageNo), "", 0, "C", false, 0, "")
	})

	for i, laporan := range laporanDocs {
		// Tambahkan halaman baru
		pdf.AddPage()
		pdf.SetFont("Arial", "UB", 16)
		pdf.MultiCell(
			0,                                      // Lebar: 0 berarti lebar otomatis
			10,                                     // Tinggi baris
			"Risalah Pertemuan-"+strconv.Itoa(i+1), // Teks
			"",                                     // Batas kiri
			"",                                     // Batas kanan
			false,                                  // Aligment horizontal
		)
		// Tambahkan teks ke PDF
		pdf.SetFont("Arial", "B", 12)
		pdf.MultiCell(
			0,                         // Lebar: 0 berarti lebar otomatis
			5,                         // Tinggi baris
			laporan.MeetEvent.Summary, // Teks
			"",                        // Batas kiri
			"",                        // Batas kanan
			false,                     // Aligment horizontal
		)
		pdf.SetFont("Arial", "I", 12)
		pdf.MultiCell(
			0,                          // Lebar: 0 berarti lebar otomatis
			5,                          // Tinggi baris
			"Notula: "+laporan.Petugas, // Teks
			"",                         // Batas kiri
			"",                         // Batas kanan
			false,                      // Aligment horizontal
		)
		pdf.MultiCell(
			0, // Lebar: 0 berarti lebar otomatis
			5, // Tinggi baris
			"Waktu: "+laporan.MeetEvent.Date+" ("+laporan.MeetEvent.TimeStart+" - "+laporan.MeetEvent.TimeEnd+")", // Teks
			"",    // Batas kiri
			"",    // Batas kanan
			false, // Aligment horizontal
		)
		pdf.MultiCell(
			0,                                     // Lebar: 0 berarti lebar otomatis
			5,                                     // Tinggi baris
			"Lokasi: "+laporan.MeetEvent.Location, // Teks
			"",                                    // Batas kiri
			"",                                    // Batas kanan
			false,                                 // Aligment horizontal
		)
		pdf.MultiCell(
			0,         // Lebar: 0 berarti lebar otomatis
			5,         // Tinggi baris
			"Agenda:", // Teks
			"",        // Batas kiri
			"",        // Batas kanan
			false,     // Aligment horizontal
		)
		pdf.MultiCell(
			0,              // Lebar: 0 berarti lebar otomatis
			5,              // Tinggi baris
			laporan.Solusi, // Teks
			"",             // Batas kiri
			"",             // Batas kanan
			false,          // Aligment horizontal
		)
		pdf.SetFont("Arial", "UB", 12)
		pdf.MultiCell(
			0,         // Lebar: 0 berarti lebar otomatis
			5,         // Tinggi baris
			"Risalah", // Teks
			"",        // Batas kiri
			"",        // Batas kanan
			false,     // Aligment horizontal
		)
		pdf.SetFont("Arial", "", 12)
		pdf.MultiCell(
			0,                // Lebar: 0 berarti lebar otomatis
			5,                // Tinggi baris
			laporan.Komentar, // Teks
			"",               // Batas kiri
			"",               // Batas kanan
			false,            // Aligment horizontal
		)
	}

	// Simpan PDF ke file sementara
	tempFile := projectName
	err = pdf.OutputFileAndClose(tempFile)
	if err != nil {
		return "", err
	}

	// Baca file PDF dan konversi ke base64
	fileData, err := ioutil.ReadFile(tempFile)
	if err != nil {
		return "", err
	}

	base64Str = base64.StdEncoding.EncodeToString(fileData)

	// Hapus file sementara
	err = os.Remove(tempFile)
	if err != nil {
		return "", err
	}

	return

}

func RekapPagiHari(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	filter := bson.M{"_id": YesterdayFilter()}
	wagroupidlist, err := atdb.GetAllDistinctDoc(config.Mongoconn, filter, "project.wagroupid", "pushrepo")
	if err != nil {
		resp.Info = "Gagal Query Distincs project.wagroupid"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	for _, gid := range wagroupidlist { //iterasi di setiap wa group
		// Type assertion to convert any to string
		groupID, ok := gid.(string)
		if !ok {
			resp.Info = "wagroupid is not a string"
			resp.Response = "wagroupid is not a string"
			at.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
		msg, err := GenerateRekapMessageKemarinPerWAGroupID(config.Mongoconn, groupID)
		if err != nil {
			resp.Info = "Gagal Membuat Rekapitulasi perhitungan per wa group id"
			resp.Response = err.Error()
			at.WriteJSON(respw, http.StatusExpectationFailed, resp)
			return
		}
		dt := &whatsauth.TextMessage{
			To:       groupID,
			IsGroup:  true,
			Messages: msg,
		}
		_, resp, err = atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
		if err != nil {
			resp.Info = "Tidak berhak"
			resp.Response = err.Error()
			at.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
	}

}
