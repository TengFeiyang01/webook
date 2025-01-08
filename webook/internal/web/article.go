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
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
	ijwt "webook/webook/internal/web/jwt"
	"webook/webook/pkg/ginx"
	"webook/webook/pkg/logger"
)

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc service.InteractiveService
	biz      string
	l        logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1, interSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		l:        l,
		biz:      "article",
		interSvc: interSvc,
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
	err := h.interSvc.Collect(ctx, h.biz, req.Id, req.Cid, uc.Uid)
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
		err = h.interSvc.Like(ctx, h.biz, req.Id, uc.Uid)
	} else {
		err = h.interSvc.CancelLike(ctx, h.biz, req.Id, uc.Uid)
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
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("failed to find user's session message")
		return
	}
	err := h.svc.WithDraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
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
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("failed to find user's session message")
		return
	}
	id, err := h.svc.Save(ctx, req.toDomain(claims.Uid))
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
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("failed to find user's session message")
		return
	}
	id, err := h.svc.Publish(ctx, req.toDomain(claims.Uid))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "找不到用户",
		})
		h.l.Error("failed to publish article, used not found")
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
		Data: id,
	})
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	res, err := h.svc.List(ctx, uc.Uid, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "system error",
		}, nil
	}
	// 列表页不显示全文，只显示“摘要”
	return ginx.Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Content:  src.Content,
					Abstract: src.Abstract(),
					Author:   src.Author.Name,
					Ctime:    src.Ctime.Format(time.DateTime),
					Utime:    src.Utime.Format(time.DateTime),
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
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
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
	return ginx.Result{
		Data: ArticleVO{
			Id:       art.Id,
			Title:    art.Title,
			Abstract: art.Abstract(),
			Content:  art.Content,
			Status:   art.Status.ToUint8(),
			Ctime:    art.Ctime.Format(time.DateTime),
			Utime:    art.Utime.Format(time.DateTime),
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
		art, err = h.svc.GetPublishedById(ctx, id, usr.Uid)
		return err
	})
	if err := eg.Wait(); err != nil {
		h.l.Error("failed to get published article", logger.Error(err))
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

	var intr domain.Interactive
	eg.Go(func() error {
		intr, err = h.interSvc.Get(ctx, h.biz, id, usr.Uid)
		return err
	})
	if err := eg.Wait(); err != nil {
		h.l.Error("failed to get interactive article", logger.Error(err))
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "system error",
		}, err
	}

	// 增加阅读计数
	go func() {
		if err := h.interSvc.IncrReadCnt(ctx, h.biz, art.Id); err != nil {
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
			LikeCnt:    intr.LikeCnt,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,
		},
	}, nil
}
