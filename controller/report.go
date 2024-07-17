package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/report"
	"github.com/gocroot/helper/whatsauth"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func GetYesterdayDistincWAGroup(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	filter := bson.M{"_id": report.YesterdayFilter()}
	wagroupidlist, err := atdb.GetAllDistinctDoc(config.Mongoconn, filter, "project.wagroupid", "pushrepo")
	if err != nil {
		resp.Info = "Gagal Query Distincs project.wagroupid"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	for _, wagroupid := range wagroupidlist {
		// Type assertion to convert any to string
		groupID, ok := wagroupid.(string)
		if !ok {
			resp.Info = "wagroupid is not a string"
			resp.Response = "wagroupid is not a string"
			at.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
		//kirim report ke group
		dt := &whatsauth.TextMessage{
			To:       groupID,
			IsGroup:  true,
			Messages: report.GetDataRepoMasukHariIni(config.Mongoconn, groupID) + "\n" + report.GetDataLaporanMasukHariini(config.Mongoconn, groupID),
		}
		_, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
		if err != nil {
			resp.Info = "Tidak berhak"
			resp.Response = err.Error()
			at.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
	}
	at.WriteJSON(respw, http.StatusOK, resp)
}

func GetReportHariIni(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	//kirim report ke group
	dt := &whatsauth.TextMessage{
		To:       "6281313112053-1492882006",
		IsGroup:  true,
		Messages: report.GetDataRepoMasukHarian(config.Mongoconn) + "\n" + report.GetDataLaporanMasukHarian(config.Mongoconn),
	}
	_, resp, err := atapi.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		at.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	at.WriteJSON(respw, http.StatusOK, resp)
}

func ApproveBimbingan(w http.ResponseWriter, r *http.Request) {
	noHp := r.Header.Get("nohp")
	if noHp == "" {
		http.Error(w, "No valid phone number found", http.StatusForbidden)
		return
	}

	var requestData struct {
		NIM   string `json:"nim"`
		Topik string `json:"topik"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.NIM == "" || requestData.Topik == "" {
		http.Error(w, "Invalid request body or NIM/Topik not provided", http.StatusBadRequest)
		return
	}

	// Get the API URL from the database
	var conf model.Config
	err = config.Mongoconn.Collection("config").FindOne(context.TODO(), bson.M{"phonenumber": "62895601060000"}).Decode(&conf)
	if err != nil {
		http.Error(w, "mohon maaf ada kesalahan dalam pengambilan config di database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the request body
	requestBody, err := json.Marshal(map[string]string{
		"nim":   requestData.NIM,
		"topik": requestData.Topik,
	})
	if err != nil {
		http.Error(w, "Gagal membuat request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create and send the HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", conf.ApproveBimbinganURL, bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, "Gagal membuat request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("nohp", noHp)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Gagal mengirim request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Gagal Approve Bimbingan, status code: %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}

	var responseMap map[string]string
	err = json.NewDecoder(resp.Body).Decode(&responseMap)
	if err != nil {
		http.Error(w, "Gagal memproses response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kurangi poin berdasarkan nomor telepon yang ada di response
	phonenumber := responseMap["no_hp"]
	_, err = report.KurangPoinUserbyPhoneNumber(config.Mongoconn, phonenumber, 13.0)
	if err != nil {
		http.Error(w, "Gagal mengurangi poin: "+err.Error(), http.StatusInternalServerError)
		return
	}

	at.WriteJSON(w, http.StatusOK, responseMap)
}
