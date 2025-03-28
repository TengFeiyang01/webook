package web

import (
	"errors"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	artv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/article/v1"
	intrv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/intr/v1"
	"github.com/TengFeiyang01/webook/webook/article/domain"
	ijwt "github.com/TengFeiyang01/webook/webook/internal/web/jwt"
	"github.com/TengFeiyang01/webook/webook/pkg/ginx"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc      artv1.ArticleServiceClient
	interSvc intrv1.InteractiveServiceClient
	biz      string
	l        logger.LoggerV1
}

func NewArticleHandler(svc artv1.ArticleServiceClient, l logger.LoggerV1, intrSvc intrv1.InteractiveServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: intrSvc,
		l:        l,
		biz:      "art",
	}
}
func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/withdraw", h.WithDraw)
	g.POST("/publish", h.Publish)
	// 创作者的查询接口
	// 这个是获取数据的接口，理论上来说（遵循 RESTful 规范），应该是 GET 方法
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](h.List))
	g.POST("/detail/:id", ginx.WrapToken[ijwt.UserClaims](h.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapToken[ijwt.UserClaims](h.PubDetail))
	// 点赞和取消点赞都复用这个接口
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq, ijwt.UserClaims](h.Like))
	pub.POST("/collect", ginx.WrapBodyAndToken[CollectReq, ijwt.UserClaims](h.Collect))
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := h.interSvc.Collect(ctx, &intrv1.CollectRequest{
		Biz:   h.biz,
		BizId: req.Id,
		Cid:   req.Cid,
		Uid:   uc.Uid,
	})
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "system error",
		}, err
	}
	return ginx.Result{Code: http.StatusOK}, nil
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = h.interSvc.Like(ctx, &intrv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	} else {
		_, err = h.interSvc.CancelLike(ctx, &intrv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.Uid,
		})
	}

	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "system error",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) WithDraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("user")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("failed to find user's session message")
		return
	}
	_, err := h.svc.WithDraw(ctx, &artv1.WithDrawRequest{
		Art: &artv1.Article{
			Id: req.Id,
			Author: &artv1.Author{
				Id: claims.Uid,
			},
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Info("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg:  "OK",
		Data: req.Id,
	})
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {

	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 检测输入、跳过这一步
	// 调用 svc 的代码
	c := ctx.MustGet("user")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("failed to find user's session message")
		return
	}
	resp, err := h.svc.Save(ctx, &artv1.SaveRequest{
		Art: &artv1.Article{
			Id:      req.Id,
			Title:   req.Title,
			Content: req.Content,
			Author: &artv1.Author{
				Id: claims.Uid,
			},
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Info("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg:  "OK",
		Data: resp.GetId(),
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("user")
	claims, ok := c.(ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("failed to find user's session message")
		return
	}
	resp, err := h.svc.Publish(ctx, &artv1.PublishRequest{
		Art: &artv1.Article{
			Id:      req.Id,
			Title:   req.Title,
			Content: req.Content,
			Author: &artv1.Author{
				Id: claims.Uid,
			},
		},
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "找不到用户",
		})
		h.l.Error("failed to publish art, used not found")
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Info("发布帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg:  "OK",
		Data: resp.GetId(),
	})
}

func (h *ArticleHandler) Abstract(content string) string {
	cs := []rune(content)
	if len(cs) < 100 {
		return content
	}
	return content[:100]
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.svc.List(ctx, &artv1.ListRequest{
		Id:     uc.Uid,
		Offset: int32(req.Offset),
		Limit:  int32(req.Limit),
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "system error",
		}, nil
	}
	// 列表页不显示全文，只显示“摘要”
	return ginx.Result{
		Data: slice.Map[*artv1.Article, ArticleVO](resp.Arts,
			func(idx int, src *artv1.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Content:  src.Content,
					Abstract: h.Abstract(src.Content),
					Author:   src.Author.Name,
					Ctime:    time.UnixMilli(src.Ctime).Format(time.DateTime),
					Utime:    time.UnixMilli(src.Utime).Format(time.DateTime),
				}
			}),
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context, usr ijwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		//h.l.Error("id input error", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	resp, err := h.svc.GetById(ctx, &artv1.GetByIdRequest{Id: id})
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "system error",
		}, err
	}
	// 不借助数据库查询来判断
	if resp.Art.Author.Id != usr.Uid {
		return ginx.Result{
			Code: http.StatusBadRequest,
			Msg:  "input error",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.Uid)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:       resp.Art.Id,
			Title:    resp.Art.Title,
			Abstract: h.Abstract(resp.Art.Content),
			Content:  resp.Art.Content,
			Status:   domain.ArticleStatus(resp.Art.Status).ToUint8(),
			Ctime:    time.UnixMilli(resp.Art.Ctime).Format(time.DateTime),
			Utime:    time.UnixMilli(resp.Art.Utime).Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, usr ijwt.UserClaims) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		h.l.Error("id input error", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}

	var art domain.Article
	var eg errgroup.Group
	// 读文章本体
	eg.Go(func() error {
		resp, err := h.svc.GetPubById(ctx, &artv1.GetPubByIdRequest{Id: id, Uid: usr.Uid})
		art = domain.Article{
			Id:      resp.Art.Id,
			Title:   resp.Art.Title,
			Content: resp.Art.Content,
			Status:  domain.ArticleStatus(resp.Art.Status),
			Ctime:   time.UnixMilli(resp.Art.Ctime),
			Utime:   time.UnixMilli(resp.Art.Utime),
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		h.l.Error("failed to get published art", logger.Error(err))
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "system error",
		}, err
	}

	// 不借助数据库查询来判断
	if art.Author.Id != usr.Uid {
		return ginx.Result{
			Code: http.StatusBadRequest,
			Msg:  "input error",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.Uid)
	}

	// 要在这里获得文章的计数

	var resp *intrv1.GetResponse
	eg.Go(func() error {
		resp, err = h.interSvc.Get(ctx, &intrv1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   usr.Uid,
		})
		return err
	})
	if err := eg.Wait(); err != nil {
		h.l.Error("failed to get interactive art", logger.Error(err))
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "system error",
		}, err
	}

	// 增加阅读计数
	go func() {
		if _, err := h.interSvc.IncrReadCnt(ctx, &intrv1.IncrReadCntRequest{
			Biz:   h.biz,
			BizId: art.Id,
		}); err != nil {
			h.l.Error("incrReadCnt failed", logger.Int64("aid", art.Id), logger.Error(err))
		}
	}()

	return ginx.Result{
		Data: ArticleVO{
			Id:         art.Id,
			Title:      art.Title,
			Status:     art.Status.ToUint8(),
			Content:    art.Content,
			Author:     art.Author.Name,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			LikeCnt:    resp.Intr.LikeCnt,
			ReadCnt:    resp.Intr.ReadCnt,
			CollectCnt: resp.Intr.CollectCnt,
			Liked:      resp.Intr.Liked,
			Collected:  resp.Intr.Collected,
		},
	}, nil
}
