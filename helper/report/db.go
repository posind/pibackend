package report

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetPreviousPoin(db *mongo.Database, collection string, ObjectID primitive.ObjectID, arrayname string, filterCondition bson.M) (previousPoin float64, err error) {
	// Membuat filter untuk menemukan dokumen dengan ID dan elemen array yang sesuai dengan kondisi filter
	filter := bson.M{
		"_id": ObjectID,
	}

	// Menambahkan kondisi filter untuk elemen dalam array
	for key, value := range filterCondition {
		filter[fmt.Sprintf("%s.%s", arrayname, key)] = value
	}

	// Membuat projection untuk hanya mengambil elemen array yang sesuai
	projection := bson.M{
		fmt.Sprintf("%s.$", arrayname): 1,
	}

	var result bson.M
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mengambil dokumen dari koleksi
	err = db.Collection(collection).FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		return
	}

	// Mendapatkan elemen array yang sesuai
	members := result[arrayname].(bson.A)
	if len(members) > 0 {
		member := members[0].(bson.M)
		previousPoin = member["poin"].(float64)
	} else {
		err = fmt.Errorf("member not found")
	}

	return
}
