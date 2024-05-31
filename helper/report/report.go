package report

import (
	"context"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PushRank struct {
	Username    string
	TotalCommit int
	Repos       map[string]int
}

func GetDataLaporanMasukHarian(db *mongo.Database) (msg string) {
	msg += "*Jumlah Laporan Hari Ini :*\n"
	ranklist := GetRankDataLayananHarian(db, TodayFilter())
	for i, data := range ranklist {
		msg += strconv.Itoa(i+1) + ". " + data.Username + " : " + strconv.Itoa(data.TotalCommit) + "\n"
	}

	return
}
func GetRankDataLayananHarian(db *mongo.Database, filterhari bson.M) (ranklist []PushRank) {
	pushrepo := db.Collection("uxlaporan")
	// Create filter to query data for today
	filter := bson.M{"_id": filterhari}
	usernamelist, _ := atdb.GetAllDistinctDoc(db, filter, "petugas", "uxlaporan")
	//ranklist := []PushRank{}
	for _, username := range usernamelist {
		filter := bson.M{"petugas": username, "_id": filterhari}
		// Query the database
		var pushdata []model.Laporan
		cur, err := pushrepo.Find(context.Background(), filter)
		if err != nil {
			return
		}
		if err = cur.All(context.Background(), &pushdata); err != nil {
			return
		}
		defer cur.Close(context.Background())
		if len(pushdata) > 0 {
			ranklist = append(ranklist, PushRank{Username: username.(string), TotalCommit: len(pushdata)})
		}
	}
	sort.SliceStable(ranklist, func(i, j int) bool {
		return ranklist[i].TotalCommit > ranklist[j].TotalCommit
	})
	return
}

func GetDataRepoMasukKemarinBukanLibur(db *mongo.Database) (msg string) {
	msg += "*Laporan Jumlah Push Repo Hari Ini :*\n"
	pushrepo := db.Collection("pushrepo")
	// Create filter to query data for today
	filter := bson.M{"_id": YesterdayNotLiburFilter()}
	usernamelist, _ := atdb.GetAllDistinctDoc(db, filter, "username", "pushrepo")
	for _, username := range usernamelist {
		filter := bson.M{"username": username, "_id": YesterdayNotLiburFilter()}
		// Query the database
		var pushdata []model.PushReport
		cur, err := pushrepo.Find(context.Background(), filter)
		if err != nil {
			return
		}
		if err = cur.All(context.Background(), &pushdata); err != nil {
			return
		}
		defer cur.Close(context.Background())
		if len(pushdata) > 0 {
			msg += "*" + username.(string) + " : " + strconv.Itoa(len(pushdata)) + "*\n"
			for j, push := range pushdata {
				msg += strconv.Itoa(j+1) + ". " + strings.TrimSpace(push.Message) + "\n"

			}
		}
	}
	return
}

func GetDataRepoMasukHarian(db *mongo.Database) (msg string) {
	msg += "*Laporan Jumlah Push Repo Hari Ini :*\n"
	pushrepo := db.Collection("pushrepo")
	// Create filter to query data for today
	filter := bson.M{"_id": TodayFilter()}
	usernamelist, _ := atdb.GetAllDistinctDoc(db, filter, "username", "pushrepo")
	for _, username := range usernamelist {
		filter := bson.M{"username": username, "_id": TodayFilter()}
		// Query the database
		var pushdata []model.PushReport
		cur, err := pushrepo.Find(context.Background(), filter)
		if err != nil {
			return
		}
		if err = cur.All(context.Background(), &pushdata); err != nil {
			return
		}
		defer cur.Close(context.Background())
		if len(pushdata) > 0 {
			msg += "*" + username.(string) + " : " + strconv.Itoa(len(pushdata)) + "*\n"
			for j, push := range pushdata {
				msg += strconv.Itoa(j+1) + ". " + strings.TrimSpace(push.Message) + "\n"

			}
		}
	}
	return
}

func GetRankDataRepoMasukHarian(db *mongo.Database, filterhari bson.M) (ranklist []PushRank) {
	pushrepo := db.Collection("pushrepo")
	// Create filter to query data for today
	filter := bson.M{"_id": filterhari}
	usernamelist, _ := atdb.GetAllDistinctDoc(db, filter, "username", "pushrepo")
	//ranklist := []PushRank{}
	for _, username := range usernamelist {
		filter := bson.M{"username": username, "_id": filterhari}
		cur, err := pushrepo.Find(context.Background(), filter)
		if err != nil {
			log.Println("Failed to find pushrepo data:", err)
			return
		}

		defer cur.Close(context.Background())

		repoCommits := make(map[string]int)
		for cur.Next(context.Background()) {
			var report model.PushReport
			if err := cur.Decode(&report); err != nil {
				log.Println("Failed to decode pushrepo data:", err)
				return
			}
			repoCommits[report.Repo]++
		}

		if len(repoCommits) > 0 {
			totalCommits := 0
			for _, count := range repoCommits {
				totalCommits += count
			}
			ranklist = append(ranklist, PushRank{Username: username.(string), TotalCommit: totalCommits, Repos: repoCommits})
		}
	}
	sort.SliceStable(ranklist, func(i, j int) bool {
		return ranklist[i].TotalCommit > ranklist[j].TotalCommit
	})
	return
}

func GetDateSekarang() (datesekarang time.Time) {
	// Definisi lokasi waktu sekarang
	location, _ := time.LoadLocation("Asia/Jakarta")

	t := time.Now().In(location) //.Truncate(24 * time.Hour)
	datesekarang = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return
}

func TodayFilter() bson.M {
	return bson.M{
		"$gte": primitive.NewObjectIDFromTimestamp(GetDateSekarang()),
		"$lt":  primitive.NewObjectIDFromTimestamp(GetDateSekarang().Add(24 * time.Hour)),
	}
}

func YesterdayNotLiburFilter() bson.M {
	return bson.M{
		"$gte": primitive.NewObjectIDFromTimestamp(GetDateKemarinBukanHariLibur()),
		"$lt":  primitive.NewObjectIDFromTimestamp(GetDateKemarinBukanHariLibur().Add(24 * time.Hour)),
	}
}

func Yesterday() bson.M {
	return bson.M{
		"$gte": primitive.NewObjectIDFromTimestamp(GetDateKemarin()),
		"$lt":  primitive.NewObjectIDFromTimestamp(GetDateKemarin().Add(24 * time.Hour)),
	}
}

func GetDateKemarinBukanHariLibur() (datekemarinbukanlibur time.Time) {
	// Definisi lokasi waktu sekarang
	location, _ := time.LoadLocation("Asia/Jakarta")
	n := -1
	t := time.Now().AddDate(0, 0, n).In(location) //.Truncate(24 * time.Hour)
	for HariLibur(t) {
		n -= 1
		t = time.Now().AddDate(0, 0, n).In(location)
	}

	datekemarinbukanlibur = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return
}

func GetDateKemarin() (datekemarin time.Time) {
	// Definisi lokasi waktu sekarang
	location, _ := time.LoadLocation("Asia/Jakarta")
	n := -1
	t := time.Now().AddDate(0, 0, n).In(location) //.Truncate(24 * time.Hour)
	datekemarin = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return
}

func HariLibur(thedate time.Time) (libur bool) {
	wekkday := thedate.Weekday()
	inhari := int(wekkday)
	if inhari == 0 || inhari == 6 {
		libur = true
	}
	tglskr := thedate.Format("2006-01-02")
	tgl := int(thedate.Month())
	urltarget := "https://dayoffapi.vercel.app/api?month=" + strconv.Itoa(tgl)
	hasil, _ := helper.Get[[]NewLiburNasional](urltarget)
	for _, v := range hasil {
		if v.Tanggal == tglskr {
			libur = true
		}
	}
	return
}

func Last3DaysFilter() bson.M {
	tigaHariLalu := GetDateSekarang().Add(-72 * time.Hour) // 3 * 24 hours
	now := GetDateSekarang()
	return bson.M{
		"$gte": primitive.NewObjectIDFromTimestamp(tigaHariLalu),
		"$lt":  primitive.NewObjectIDFromTimestamp(now),
	}
}
