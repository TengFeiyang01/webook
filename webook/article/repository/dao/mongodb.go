package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m *MongoDBArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) Upsert(ctx context.Context, art PublishedArticleV1) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	_, err := m.liveCol.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))
	return err
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	filter := bson.D{bson.E{Key: "author_id", Value: uid}}
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{"ctime", -1}}) // 按 ctime 倒序排列
	cur, err := m.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var articles []Article
	for cur.Next(ctx) {
		var art Article
		if err := cur.Decode(&art); err != nil {
			return nil, err
		}
		articles = append(articles, art)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return articles, nil
}

func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	filter := bson.D{bson.E{Key: "id", Value: id}}
	var art Article
	err := m.col.FindOne(ctx, filter).Decode(&art)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return art, errors.New("art not found")
		}
		return art, err
	}
	return art, nil
}

func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (Article, error) {
	filter := bson.D{bson.E{Key: "id", Value: id}}
	var art Article
	err := m.liveCol.FindOne(ctx, filter).Decode(&art)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return art, errors.New("published art not found")
		}
		return art, err
	}
	return art, nil
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, &art)
	return art.Id, err
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		// 创作者不对，说明有人在瞎搞
		return errors.New("ID 不对或者创作者不对")
	}
	return nil
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now

	// liveCol 是 INSERT or Update 语义
	//filter := bson.D{bson.E{Key: "id", Value: id},
	//	bson.E{Key: "author_id", Value: art.AuthorId}}
	filter := bson.M{"id": art.Id}
	//set := bson.D{bson.E{Key: "$set", Value: art},
	//	bson.E{Key: "$setOnInsert",
	//		Value: bson.D{bson.E{Key: "ctime", Value: now}}}}
	updateV1 := bson.M{
		"$set":         PublishedArticle(art),
		"$setOnInsert": bson.M{"ctime": now},
	}

	_, err = m.liveCol.UpdateOne(ctx, filter,
		updateV1,
		//set,
		options.Update().SetUpsert(true))
	return id, err
}

func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	filter := bson.D{bson.E{Key: "id", Value: id},
		bson.E{Key: "author_id", Value: uid}}
	sets := bson.D{bson.E{Key: "$set",
		Value: bson.D{bson.E{Key: "status", Value: status}}}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("ID 不对或者创作者不对")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, sets)
	return err
}

var _ ArticleDAO = &MongoDBArticleDAO{}

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) *MongoDBArticleDAO {
	return &MongoDBArticleDAO{
		node:    node,
		liveCol: mdb.Collection("published_articles"),
		col:     mdb.Collection("articles"),
	}
}

func ToUpdates(vals map[string]any) bson.M {
	return vals
}

func ToFilter(vals map[string]any) (res bson.D) {
	for k, v := range vals {
		res = append(res, bson.E{Key: k, Value: v})
	}
	return
}

func Set(vals map[string]any) bson.M {
	return bson.M{"$set": ToUpdates(vals)}
}

func Upsert(vals map[string]any) bson.M {
	return bson.M{"$upsert": Set(vals)}
}
