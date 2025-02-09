package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"webook/webook/internal/repository/dao/article"

	"net/http/httptest"
	"testing"
	"time"
	"webook/webook/internal/integration/startup"
	ijwt "webook/webook/internal/web/jwt"
)

type ArticleMongoDBHandlerSuite struct {
	suite.Suite
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
	server  *gin.Engine
}

func (s *ArticleMongoDBHandlerSuite) SetupSuite() {
	s.mdb = startup.InitMongoDB()
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	hdl := startup.InitArticleHandler(article.NewMongoDBArticleDAO(s.mdb, node))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", ijwt.UserClaims{
			Uid: 123,
		})
	})
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	hdl.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleMongoDBHandlerSuite) TestArticle_Publish() {
	t := s.T()

	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		req   Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "hello，你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, uint8(2), art.Status)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				var publishedArt article.PublishedArticle

				err = s.liveCol.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).
					Decode(&publishedArt)
				assert.NoError(t, err)
				t.Log(publishedArt)
				assert.Equal(t, "hello，你好", publishedArt.Title)
				assert.Equal(t, "随便试试", publishedArt.Content)
				assert.Equal(t, int64(123), publishedArt.AuthorId)
				assert.Equal(t, uint8(2), publishedArt.Status)
				//assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
				Msg:  "OK",
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 模拟已经存在的帖子
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   1,
					Utime:    234,
					AuthorId: 123,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{"id", 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, uint8(2), art.Status)
				assert.Equal(t, int64(123), art.AuthorId)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)
				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{"id", 2}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, uint8(2), publishedArt.Status)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
				Msg:  "OK",
			},
		},
		{
			"更新帖子，并且重新发表",
			func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   1,
					Utime:    234,
					AuthorId: 123,
				}
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{"id", 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, uint8(2), art.Status)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)

				var part article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{"id", 3}}).Decode(&part)
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", part.Title)
				assert.Equal(t, "新的内容", part.Content)
				assert.Equal(t, int64(123), part.AuthorId)
				assert.Equal(t, uint8(2), part.Status)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.True(t, part.Utime > 234)
			},
			Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			200,
			Result[int64]{
				Data: 3,
				Msg:  "OK",
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				art := article.Article{
					Id:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					Status:  1,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 789,
				}
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   2,
					Utime:    234,
					AuthorId: 789,
				}
				_, err = s.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 更新应该是失败了，数据没有发生变化
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{"id", 4}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(234), art.Utime)
				assert.Equal(t, uint8(1), art.Status)
				assert.Equal(t, int64(789), art.AuthorId)

				var part article.PublishedArticle
				// 数据没有变化
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{"id", 4}}).Decode(&part)
				assert.NoError(t, err)
				assert.Equal(t, "我的标题", part.Title)
				assert.Equal(t, "我的内容", part.Content)
				assert.Equal(t, int64(789), part.AuthorId)
				assert.Equal(t, uint8(2), part.Status)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.Equal(t, int64(234), part.Utime)
			},
			req: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.NoError(t, err)
			if tc.wantResult.Data > 0 {
				// 你只能断定有 ID
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}

func (s *ArticleMongoDBHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		// 前端传过来，肯定是一个 JSON
		art Article

		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "新建帖子",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				// 你要验证，保存到了数据库里面
				var art article.Article
				err := s.col.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				assert.Equal(t, article.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   1,
				}, art)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 我希望你的 ID 是 1
				Data: 1,
			},
		},
		{
			name: "修改帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 假装数据库已经有这个帖子
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:       11,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					// 假设这是一个已经发表了的帖子
					Status: 2,
					Ctime:  456,
					Utime:  789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 你要验证，保存到了数据库里面
				var art article.Article

				err := s.col.FindOne(ctx, bson.D{bson.E{"id", 11}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Utime > 789)
				art.Utime = 0
				assert.Equal(t, article.Article{
					Id:       11,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					// 更新之后，是未发表状态
					Status: 1,
					Ctime:  456,
				}, art)
			},
			art: Article{
				Id:      11,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 我希望你的 ID 是 11
				Data: 11,
			},
		},
		{
			name: "修改帖子-别人的帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 假装数据库已经有这个帖子
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:      22,
					Title:   "我的标题",
					Content: "我的内容",
					// 模拟别人
					AuthorId: 1024,
					Status:   2,
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 你要验证，保存到了数据库里面
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{"id", 22}}).
					Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, article.Article{
					Id:       22,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 1024,
					Status:   2,
					Ctime:    456,
					Utime:    789,
				}, art)
			},
			art: Article{
				Id:      22,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg: "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			// 准备Req和记录的 recorder
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit",
				bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			// 执行
			s.server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			if tc.wantRes.Data > 0 {
				// 你只能断定有 ID
				assert.True(t, res.Data > 0)
			}
		})
	}
}

func (s *ArticleMongoDBHandlerSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.liveCol.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func TestArticleMongoDBHandler(t *testing.T) {
	suite.Run(t, &ArticleMongoDBHandlerSuite{})
}
