package report

import (
	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Fungsi untuk membuat filter
func CreateFilterMeetingYesterday(projectName string) bson.M {
	startOfDay, endOfDay := atdb.GetYesterdayStartEnd()
	return bson.M{
		"_id": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
		"project.name": projectName,
		"meetid":       bson.M{"$ne": primitive.NilObjectID}, // Kondisi MeetID tidak kosong
	}
}
