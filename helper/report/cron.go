package report

import (
	"fmt"
	"net/http"

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

func RekapMeetingKemarin(db *mongo.Database, projectName string) (err error) {
	filter := CreateFilterMeetingYesterday(projectName)
	laporanDocs, err := atdb.GetAllDoc[[]Laporan](db, "uxlaporan", filter) //CreateFilterMeetingYesterday(projectName)
	fmt.Println(len(laporanDocs))
	if err != nil {
		return
	}
	// Buat PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)

	for _, laporan := range laporanDocs {
		// Tambahkan halaman baru
		pdf.AddPage()

		// Tambahkan teks ke PDF
		pdf.MultiCell(
			0,                // Lebar: 0 berarti lebar otomatis
			10,               // Tinggi baris
			laporan.Komentar, // Teks
			"",               // Batas kiri
			"",               // Batas kanan
			false,            // Aligment horizontal
		)
	}

	// Simpan PDF ke file
	err = pdf.OutputFileAndClose("output.pdf")
	if err != nil {
		return
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
