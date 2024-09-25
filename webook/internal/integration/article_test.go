package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/webook/internal/integration/startup"
	"webook/webook/internal/repository/dao"
	ijwt "webook/webook/internal/web/jwt"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	// 在所有测试执行之前，初始化
	//s.server = startup.InitWebServ er()
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	s.db = startup.InitTestDB()
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRoutes(s.server)
}

// TearDownTest 每一个测试都会执行
func (s *ArticleTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE articles")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		// 集成测试，准备数据
		before func(t *testing.T)
		// 集成测试 验证数据
		after func(t *testing.T)
		// 预期的输入
		art Article

		// HTTP 响应码
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// 验证数据库
				var art dao.Article
				err := s.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "my Title",
					Content:  "my Content",
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Title:   "my Title",
				Content: "my Content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
				Msg:  "OK",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构造请求
			// 执行
			// 验证结果
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
			tc.after(t)
		})
	}
}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("hello, 这是测试套件")
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
