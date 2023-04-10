package admin

import (
	"net/http"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

func IndexNewsfeed(ctx iris.Context) {
	newsfeed := []models.Newsfeed{}
	if err := repository.NewRepository(&models.Newsfeed{}).FindMany(&newsfeed); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   newsfeed,
	})
}

func ViewNewsfeed(ctx iris.Context) {
	newsfeed := &models.Newsfeed{}
	if ok := repository.NewRepository(newsfeed).
		FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Not found",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   newsfeed,
	})
}

func CreateNewsfeed(ctx iris.Context) {
	data := &models.NewsfeedForm{}
	if err := ctx.ReadJSON(data); !logger.HandleError(err) {
		ctx.JSON(validation.Errors(err))
		return
	}

	slug, err := models.UniqueSlug(&models.Newsfeed{}, list.Slug(data.Title))
	if !logger.HandleError(err) {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	user, _ := getUser(ctx)
	newsfeed := &models.Newsfeed{
		Id:      uuid.New(),
		UserId:  user.Id,
		Slug:    slug,
		Title:   data.Title,
		Content: data.Content,
	}

	if err := repository.NewRepository(newsfeed).Create(); !logger.HandleError(err) {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Created successfully",
	})
}

func UpdateNewsfeed(ctx iris.Context) {
	data := &models.NewsfeedUpdateForm{}
	ctx.ReadJSON(data)
	newsfeed := &models.Newsfeed{}
	if ok := repository.NewRepository(newsfeed).FindOne("slug = ?", data.Slug); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Not found",
		})
		return
	}

	if data.Title != "" {
		newsfeed.Title = data.Title
	}
	if data.Content != "" {
		newsfeed.Content = data.Content
	}

	if err := repository.NewRepository(newsfeed).Save(); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Updated successfully",
	})
}

func DeleteNewsfeed(ctx iris.Context) {
	newsfeed := &models.Newsfeed{}
	if ok := repository.NewRepository(newsfeed).FindOne("slug = ?", ctx.URLParam("slug")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "not found",
		})
		return
	}

	if err := repository.NewRepository(&models.Mock{}).
		Delete(newsfeed); !logger.HandleError(err) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "delete failed",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Deleted successfully",
	})
}
