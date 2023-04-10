package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

func IndexPodcast(ctx iris.Context) {
	podcasts := []models.Podcast{}
	if err := repository.NewRepository(&models.Podcast{}).FindMany(&podcasts); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   podcasts,
	})
}

func ViewPodcast(ctx iris.Context) {
	podcast := &models.Podcast{}
	if ok := repository.NewRepository(podcast).
		FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Podcast not found",
		})
		return
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   podcast,
	})
}

func CreatePodcast(ctx iris.Context) {
	file, h, err := ctx.FormFile("file")
	if !logger.HandleError(err) {
		return
	}
	fmt.Println(h.Filename, h.Size, h.Header)
	data := &models.PodcastForm{}
	if err := ctx.ReadForm(data); !logger.HandleError(err) {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(validation.Errors(err))
		return
	}

	objective := &models.Objective{}
	{
		fmt.Println(data.Objective)
		if err := database.UseDB("core").First(objective, "id = ?", data.Objective).Error; err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(apiResponse{
				"status":  "failed",
				"message": "objective not found",
			})
			return
		}
	}
	podcast := &models.Podcast{
		Id:          uuid.New(),
		SubjectId:   uint(objective.SubjectId),
		ObjectiveId: data.Objective,
		Title:       data.Title,
	}
	out, err := os.Create(fmt.Sprintf("storage/podcast/%s%s", podcast.Id.String(), filepath.Ext(h.Filename)))
	if !logger.HandleError(err) {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "unable to save image",
		})
		return
	}
	_, err = io.Copy(out, file)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "unable to save image",
		})
		return
	}
	podcast.Url = out.Name()

	if err := repository.NewRepository(podcast).Create(); !logger.HandleError(err) {
		ctx.JSON(internalServerError)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Podcast created successfully",
	})
}

func UpdatePodcast(ctx iris.Context) {
	data := &models.PodcastUpdateForm{}
	ctx.ReadJSON(data)
	podcast := &models.Podcast{}
	if ok := repository.NewRepository(podcast).FindOne("id = ?", data.Id); !ok {
		ctx.StatusCode(http.StatusNotFound)
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "Podcast not found",
		})
		return
	}

	if err := repository.NewRepository(podcast).Save(); !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Updated successfully",
	})
}

func DeletePodcast(ctx iris.Context) {
	podcast := &models.Podcast{}
	if ok := repository.NewRepository(podcast).FindOne("id = ?", ctx.URLParam("id")); !ok {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "not found",
		})
		return
	}

	if err := repository.NewRepository(&models.Mock{}).
		Delete(podcast); !logger.HandleError(err) {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "delete failed",
		})
		return
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Podcast deleted successfully",
	})
}
