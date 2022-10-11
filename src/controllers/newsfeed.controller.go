package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

type NewsFeedController struct {
	Ctx iris.Context
}

func (c *NewsFeedController) Get() {
	user, _ := getUser(c.Ctx)
	feed := []struct {
		models.Newsfeed
		Likes      uint `json:"likes"`
		Comments   uint `json:"comments"`
		Liked      bool `json:"is_liked"`
		Bookmarked bool `json:"is_bookmarked"`
	}{}
	database.UseDB("app").Table("newsfeeds as n").
		Select(`n.*, ni.likes, ni.liked, ni.bookmarked, nc.comments`).
		Joins(`LEFT JOIN 
		(SELECT 
			newsfeed_id,
			SUM(CASE WHEN liked = 1 THEN 1 ELSE 0 END) as likes,
			SUM(CASE WHEN liked = 1 AND user_id = ? THEN 1 ELSE 0 END) as liked,
			SUM(CASE WHEN is_bookmarked = 1 AND user_id = ? THEN 1 ELSE 0 END) as bookmarked
			FROM newsfeed_interactions GROUP BY newsfeed_id ) ni ON ni.newsfeed_id = n.id
		`, user.Id, user.Id).
		Joins(`LEFT JOIN 
			(SELECT newsfeed_id,
			COUNT(comment) as comments FROM newsfeed_comments
			GROUP BY newsfeed_id) nc ON nc.newsfeed_id = n.id
		`).Find(&feed)

	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   feed,
	})
}

func (c *NewsFeedController) Post() {
	data := struct {
		Slug    string `validate:"required"`
		Type    string `validate:"required"`
		Message string `validate:"required"`
	}{}
	if err := c.Ctx.ReadJSON(&data); err != nil {
		c.Ctx.StatusCode(400)
		c.Ctx.JSON(validation.Errors(err))
		return
	}

	feed := &models.Newsfeed{}
	if ok := repository.NewRepository(feed).
		FindOne("slug = ?", data.Slug); !ok {
		c.Ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "News not found",
		})
		return
	}
	user, _ := getUser(c.Ctx)
	{
		report := &models.NewsfeedReport{}
		if err := database.UseDB("app").First(report, "user_id = ? AND newsfeed_id = ?", user.Id, feed.Id).Error; err == nil {
			c.Ctx.StatusCode(http.StatusAccepted)
			c.Ctx.JSON(apiResponse{
				"status":  "success",
				"message": "You already sent a report on this news",
			})
			return
		}
	}
	if err := database.UseDB("app").Create(&models.NewsfeedReport{
		NewsfeedId: feed.Id,
		UserId:     user.Id,
		Type:       data.Type,
		Message:    data.Message,
	}).Error; err != nil {
		c.Ctx.StatusCode(500)
		c.Ctx.JSON(internalServerError)
		return
	}
	c.Ctx.JSON(apiResponse{
		"status": "success",
		"data":   "Report sent successfully",
	})
}

func NewsFeedView(ctx iris.Context) {
	type News struct {
		models.Newsfeed
		Likes      uint `json:"likes"`
		Liked      bool `json:"is_liked"`
		Bookmarked bool `json:"is_bookmarked"`
	}
	feed := News{}
	comments := []struct {
		models.NewsfeedComment
		Username string
	}{}
	user, _ := getUser(ctx)
	if err := database.UseDB("app").Table("newsfeeds as n").
		Select(`n.*, ni.likes, ni.liked, ni.bookmarked`).
		Joins(`LEFT JOIN
		(SELECT
			newsfeed_id,
			SUM(CASE WHEN liked = 1 THEN 1 ELSE 0 END) as likes,
			SUM(CASE WHEN liked = 1 AND user_id = ? THEN 1 ELSE 0 END) as liked,
			SUM(CASE WHEN is_bookmarked = 1 AND user_id = ? THEN 1 ELSE 0 END) as bookmarked
			FROM newsfeed_interactions GROUP BY newsfeed_id ) ni ON ni.newsfeed_id = n.id
		`, user.Id, user.Id).
		First(&feed, "slug = ?", ctx.URLParam("slug")).Error; err != nil && err == gorm.ErrRecordNotFound {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "News not found",
		})
		return
	} else if err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	database.UseDB("app").Table("newsfeed_comments as nc").
		Select("nc.*, u.username").
		Joins("LEFT JOIN users as u ON u.id = nc.user_id").
		Find(&comments, "nc.newsfeed_id = ?", feed.Id)

	ctx.JSON(apiResponse{
		"status": "success",
		"data": struct {
			News
			Comments interface{} `json:"comments"`
		}{feed, comments},
	})
}

func NewsFeedInteract(ctx iris.Context) {
	action := ctx.Params().Get("action")
	data := struct {
		Slug     string `validate:"required"`
		Comment  string
		Like     bool
		Bookmark bool
	}{}

	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(400)
		ctx.JSON(validation.Errors(err))
		return
	}

	feed := &models.Newsfeed{}
	if ok := repository.NewRepository(feed).
		FindOne("slug = ?", data.Slug); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "News not found",
		})
		return
	}

	user, _ := getUser(ctx)
	switch strings.ToLower(action) {
	case "comment":
		{
			feedComment := &models.NewsfeedComment{}
			if err := repository.NewRepository(feed).
				FindOneDst(feedComment,
					"newsfeed_id = ? AND user_id = ?",
					feed.Id,
					user.Id); err != nil && err != gorm.ErrRecordNotFound {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
				return
			} else if err == gorm.ErrRecordNotFound {
				feedComment = &models.NewsfeedComment{
					Id:         uuid.New(),
					NewsfeedId: feed.Id,
					UserId:     user.Id,
				}
				feedComment.Comment = data.Comment
				database.UseDB("app").Create(feedComment)
			} else {
				feedComment.Comment = data.Comment
				feedComment.UpdatedAt = time.Now()
				database.UseDB("app").Where("newsfeed_id = ? AND user_id = ?", feed.Id, user.Id).
					Save(feedComment)
			}
		}
	case "like":
		{
			feedInteraction := &models.NewsfeedInteraction{}
			if err := repository.NewRepository(feed).
				FindOneDst(feedInteraction,
					"newsfeed_id = ? AND user_id = ?",
					feed.Id,
					user.Id); err != nil && err != gorm.ErrRecordNotFound {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
				return
			} else if err == gorm.ErrRecordNotFound {
				feedInteraction = &models.NewsfeedInteraction{
					NewsfeedId: feed.Id,
					UserId:     user.Id,
				}
				feedInteraction.Liked = data.Like
				database.UseDB("app").Create(feedInteraction)
			} else {
				feedInteraction.Liked = data.Like
				database.UseDB("app").Where("newsfeed_id = ? AND user_id = ?", feed.Id, user.Id).
					Save(feedInteraction)
			}
		}
	case "bookmark":
		{
			feedInteraction := &models.NewsfeedInteraction{}
			if err := repository.NewRepository(feed).
				FindOneDst(feedInteraction,
					"newsfeed_id = ? AND user_id = ?",
					feed.Id,
					user.Id); err != nil && err != gorm.ErrRecordNotFound {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
				return
			} else if err == gorm.ErrRecordNotFound {
				feedInteraction = &models.NewsfeedInteraction{
					NewsfeedId: feed.Id,
					UserId:     user.Id,
				}
				feedInteraction.IsBookmarked = data.Bookmark
				database.UseDB("app").Create(feedInteraction)
			} else {
				feedInteraction.IsBookmarked = data.Bookmark
				database.UseDB("app").Where("newsfeed_id = ? AND user_id = ?", feed.Id, user.Id).
					Save(feedInteraction)
			}
		}
	}

	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Updated successfully",
	})
}

func NewsFeedReportComment(ctx iris.Context) {
	data := struct {
		Id      string `validate:"required"`
		Type    string `validate:"required"`
		Message string `validate:"required"`
	}{}
	if err := ctx.ReadJSON(&data); err != nil {
		ctx.StatusCode(400)
		ctx.JSON(validation.Errors(err))
		return
	}

	comment := &models.NewsfeedComment{}
	if err := database.UseDB("app").First(comment, "id = ?", data.Id).Error; err != nil {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Comment not found",
		})
		return
	}
	user, _ := getUser(ctx)
	{
		report := &models.NewsfeedCommentReport{}
		if err := database.UseDB("app").First(report, "user_id = ? AND newsfeed_comment_id = ?", user.Id, comment.Id).Error; err == nil {
			ctx.StatusCode(http.StatusAccepted)
			ctx.JSON(apiResponse{
				"status":  "success",
				"message": "You already sent a report on this news",
			})
			return
		}
	}
	if err := database.UseDB("app").Create(&models.NewsfeedCommentReport{
		NewsfeedCommentId: comment.Id,
		UserId:            user.Id,
		Type:              data.Type,
		Message:           data.Message,
	}).Error; err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   "Report sent successfully",
	})
}
