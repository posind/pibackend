package atdb

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoConnect(mconn DBInfo) (db *mongo.Database, err error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mconn.DBString))
	if err != nil {
		mconn.DBString = SRVLookup(mconn.DBString)
		client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mconn.DBString))
		if err != nil {
			return
		}
	}
	db = client.Database(mconn.DBName)
	return
}

func SRVLookup(srvuri string) (mongouri string) {
	atsplits := strings.Split(srvuri, "@")
	userpass := strings.Split(atsplits[0], "//")[1]
	mongouri = "mongodb://" + userpass + "@"
	slashsplits := strings.Split(atsplits[1], "/")
	domain := slashsplits[0]
	dbname := slashsplits[1]
	//"mongodb://john:PASSWORD@gdelt-shard-00-00.n1mbb.mongodb.net:27017,gdelt-shard-00-01.n1mbb.mongodb.net:27017,gdelt-shard-00-02.n1mbb.mongodb.net:27017/DATABASE?ssl=true&authSource=admin&replicaSet=atlas-7o9d3y-shard-0"
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}
	_, srvs, err := r.LookupSRV(context.Background(), "mongodb", "tcp", domain)
	if err != nil {
		panic(err)
	}
	var srvlist string
	for _, srv := range srvs {
		srvlist += strings.TrimSuffix(srv.Target, ".") + ":" + strconv.FormatUint(uint64(srv.Port), 10) + ","
	}

	txtrecords, _ := r.LookupTXT(context.Background(), domain)
	var txtlist string
	for _, txt := range txtrecords {
		txtlist += txt
	}
	mongouri = mongouri + strings.TrimSuffix(srvlist, ",") + "/" + dbname + "?ssl=true&" + txtlist
	return
}

func GetAllDistinctDoc(db *mongo.Database, filter bson.M, fieldname, collection string) (doc []any, err error) {
	ctx := context.TODO()
	doc, err = db.Collection(collection).Distinct(ctx, fieldname, filter)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// GetAllDistinctDoc mengambil semua nilai yang berbeda dari field tertentu dalam koleksi yang diberikan
func GetAllDistinct[T any](db *mongo.Database, filter bson.M, fieldname, collection string) ([]T, error) {
	ctx := context.TODO()
	rawDoc, err := db.Collection(collection).Distinct(ctx, fieldname, filter)
	if err != nil {
		return nil, err
	}

	// Mengkonversi []interface{} ke []T
	result := make([]T, len(rawDoc))
	for i, v := range rawDoc {
		value, ok := v.(T)
		if !ok {
			return nil, fmt.Errorf("type assertion to %T failed", v)
		}
		result[i] = value
	}
	return result, nil
}

func GetRandomDoc[T any](db *mongo.Database, collection string, size uint) (result []T, err error) {
	filter := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: size}}}},
	}
	ctx := context.Background()
	cursor, err := db.Collection(collection).Aggregate(ctx, filter)
	if err != nil {
		return
	}

	err = cursor.All(ctx, &result)

	return
}

func GetAllDoc[T any](db *mongo.Database, collection string, filter bson.M) (doc T, err error) {
	ctx := context.Background()
	cur, err := db.Collection(collection).Find(ctx, filter)
	if err != nil {
		return
	}
	defer cur.Close(ctx)
	err = cur.All(ctx, &doc)
	if err != nil {
		return
	}
	return
}

func GetOneDoc[T any](db *mongo.Database, collection string, filter bson.M) (doc T, err error) {
	err = db.Collection(collection).FindOne(context.Background(), filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func DeleteOneDoc(db *mongo.Database, collection string, filter bson.M) error {
	_, err := db.Collection(collection).DeleteOne(context.Background(), filter)
	return err
}

func GetOneLatestDoc[T any](db *mongo.Database, collection string, filter bson.M) (doc T, err error) {
	opts := options.FindOne().SetSort(bson.M{"$natural": -1})
	err = db.Collection(collection).FindOne(context.TODO(), filter, opts).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func InsertOneDoc(db *mongo.Database, collection string, doc interface{}) (insertedID primitive.ObjectID, err error) {
	insertResult, err := db.Collection(collection).InsertOne(context.TODO(), doc)
	if err != nil {
		return
	}
	return insertResult.InsertedID.(primitive.ObjectID), nil
}

// With replaceOne() you can only replace the entire document,
// while updateOne() allows for updating fields. Since replaceOne() replaces the entire document - fields in the old document not contained in the new will be lost.
// With updateOne() new fields can be added without losing the fields in the old document.
func UpdateDoc(db *mongo.Database, collection string, filter bson.M, updatefield bson.M) (updateresult *mongo.UpdateResult, err error) {
	updateresult, err = db.Collection(collection).UpdateOne(context.TODO(), filter, updatefield)
	if err != nil {
		return
	}
	return
}

func ReplaceOneDoc(db *mongo.Database, collection string, filter bson.M, doc interface{}) (updatereseult *mongo.UpdateResult, err error) {
	updatereseult, err = db.Collection(collection).ReplaceOne(context.TODO(), filter, doc)
	if err != nil {
		return
	}
	return
}
