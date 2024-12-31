package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestMongoDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Println(evt.Command)
		},
	}
	opts := options.Client().
		ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)
	// 操作 client
	col := client.Database("webook").
		Collection("articles")
	defer func() {
		_, err = col.DeleteMany(ctx, bson.D{})
	}()

	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 123,
	})
	assert.NoError(t, err)

	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log("插入ID", oid)
	//filter := bson.D{bson.E{"id", 1}}
	filter := bson.M{
		"id": 1,
	}
	findRes := col.FindOne(ctx, filter)
	if errors.Is(findRes.Err(), mongo.ErrNoDocuments) {
		t.Log("没找到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var art Article
		err = findRes.Decode(&art)
		assert.NoError(t, err)
		t.Log(art)
	}

	updateFilter := bson.D{bson.E{"id", 1}}
	//set := bson.D{bson.E{Key: "$set", Value: bson.E{Key: "title", Value: "新的标题"}}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title": "新的标题",
	}}}
	updateOneRes, err := col.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	updateManyRes, err := col.UpdateMany(ctx, updateFilter,
		bson.D{bson.E{Key: "$set",
			Value: Article{Content: "新的内容"}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)
	//deleteFilter := bson.D{bson.E{"id", 3}}
	//delRes, err := col.DeleteMany(ctx, deleteFilter)
	//assert.NoError(t, err)
	//t.Log("删除文档数量", delRes.DeletedCount)

	//or := bson.A{bson.D{bson.E{"id", 1}}, bson.D{bson.E{"id", 2}}}
	or := bson.A{bson.M{"id": 1}, bson.M{"id": 2}}
	orRes, err := col.Find(ctx, bson.D{bson.E{"$or", or}})
	assert.NoError(t, err)
	var ars []Article
	err = orRes.All(ctx, &ars)
	assert.NoError(t, err)
	t.Log(ars)

	//and := bson.A{
	//	bson.D{bson.E{Key: "id", Value: 1}},
	//	bson.D{bson.E{Key: "title", Value: "新的标题"}},
	//}
	and := bson.A{bson.M{"id": 1}, bson.M{"title": "新的标题"}}
	//andRes, err := col.Find(ctx, bson.D{bson.E{"$and", and}})
	andRes, err := col.Find(ctx, bson.D{bson.E{"$and", and}})
	assert.NoError(t, err)
	err = andRes.All(ctx, &ars)
	assert.NoError(t, err)
	t.Log(ars)

	in := bson.D{bson.E{"id", bson.D{bson.E{"$in", []any{1, 2}}}}}
	//inRes, err := col.Find(ctx, in, options.Find().SetProjection(bson.M{
	//	"id":    1,
	//	"title": 1,
	//}))
	inRes, err := col.Find(ctx, in)
	assert.NoError(t, err)
	err = inRes.All(ctx, &ars)
	assert.NoError(t, err)
	t.Log(ars)

	idxRes, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"author_id": 1},
		},
	})
	assert.NoError(t, err)
	t.Log(idxRes)
}

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"`
	AuthorId int64  `bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
