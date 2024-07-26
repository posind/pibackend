package atdb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetYesterdayStartEnd() (startOfDay, endOfDay time.Time) {
	// Hitung tanggal kemarin
	loc, _ := time.LoadLocation("Asia/Jakarta")
	yesterday := time.Now().In(loc).AddDate(0, 0, -1)
	startOfDay = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, loc)
	endOfDay = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999, loc)
	return
}

// Fungsi untuk membuat filter
func CreateFilter(startOfDay, endOfDay time.Time, fieldName, fieldValue string) bson.M {
	return bson.M{
		"timestamp": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
		fieldName: fieldValue,
	}
}
