package report

import (
	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
)

// Fungsi untuk membuat filter
func CreateFilterMeetingYesterday(projectName string, ismeeting bool) bson.M {
	startOfDay, endOfDay := atdb.GetYesterdayStartEnd()
	return bson.M{
		"_id": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
		"project.name": projectName,
		"meetid":       bson.M{"$exists": ismeeting}, // Kondisi MeetID tidak kosong
	}
}
