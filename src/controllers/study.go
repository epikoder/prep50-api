package controllers

import (
	"math"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/kataras/iris/v12"
)

//++++++++++++++++++++++++++++++++++++++++++++++++
func StudySubjects(ctx iris.Context) {
	type (
		subjectForm struct {
			Exam []string
			// WithObjective bool
		}
	)
	var response map[string][]models.UserSubjectProgress = make(map[string][]models.UserSubjectProgress)

	data := &subjectForm{}
	ctx.ReadJSON(data)
	user, _ := getUser(ctx)
	session := settings.Get("exam.session", time.Now().Year())
	examSubject := map[query][]models.UserSubject{}
	{
		q := []query{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, ue.user_id, ue.session, e.name, e.status").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("ue.session = ? AND e.status = 1 AND ue.user_id = ?", session, user.Id).
			Scan(&q).Error; !logger.HandleError(err) {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}

		for _, e := range q {
			if len(data.Exam) > 0 && !list.Contains(data.Exam, e.Name) || !e.Status {
				continue
			}

			userSubjects := []models.UserSubject{}
			if err := repository.NewRepository(&models.UserSubject{}).FindMany(&userSubjects, "user_id = ? AND user_exam_id = ?", user.Id, e.Id); !logger.HandleError(err) {
				ctx.StatusCode(http.StatusInternalServerError)
				ctx.JSON(internalServerError)
				return
			}
			examSubject[e] = append(examSubject[e], userSubjects...)
		}
	}
	ids := map[string][]uint{}
	for e, us := range examSubject {
		for _, s := range us {
			ids[e.Name] = append(ids[e.Name], s.SubjectId)
		}
	}
	repo := repository.NewRepository(&models.Subject{})
	// if data.WithObjective {
	// 	repo.Preload("Objectives")
	// }
	progress := []models.UserProgress{}
	database.UseDB("app").Find(&progress, "user_id = ?", user.Id)

	objectives := []models.Objective{}
	database.UseDB("core").Find(&objectives, "id = ?", user.Id)

	for e := range examSubject {
		subjects := []models.Subject{}
		if err := repo.FindMany(&subjects, "id IN ?", ids[e.Name]); !logger.HandleError(err) {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
		for _, sub := range subjects {
			response[e.Name] = append(response[e.Name], models.UserSubjectProgress{
				Id:          sub.Id,
				Name:        sub.Name,
				Description: sub.Description,
				Progress:    models.FindSubjectProgressFromList(progress, sub.Id),
			})
		}
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   response,
	})
}

func StudyLessons(ctx iris.Context) {
	type topicForm struct {
		Subject              []int
		Objective            []int
		WithObjective        bool
		WithLesson           bool
		FilterEmptyTopic     bool
		FilterEmptyObjective bool
	}
	data := &topicForm{}
	ctx.ReadJSON(data)
	if len(data.Subject) > 0 {
		data.Subject = list.Unique(data.Subject).([]int)
	}
	if len(data.Objective) > 0 {
		data.Objective = list.Unique(data.Objective).([]int)
	}

	user, _ := getUser(ctx)
	session := settings.Get("exam.session", time.Now().Year())
	type ID struct {
		SubjectId     int
		PaymentStatus models.PaymentStatus
	}
	ids := []ID{}
	{
		if err := database.UseDB("app").
			Table("user_subjects as us").
			Select("us.subject_id").
			Joins("LEFT JOIN user_exams as ue on ue.id = us.user_exam_id").
			// Where("us.user_id = ? AND ue.session = ? AND ue.payment_status = ?", user.Id, session, models.Completed).
			Where("us.user_id = ? AND ue.session = ?", user.Id, session).
			Find(&ids).Error; !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
	}

	allowedIds := []int{}
	{
		for _, i := range ids {
			allowedIds = append(allowedIds, i.SubjectId)
		}
		if len(data.Subject) > 0 {
			tmp := allowedIds
			allowedIds = []int{}
			for _, id := range data.Subject {
				if list.Contains(tmp, id) {
					allowedIds = append(allowedIds, id)
				}
			}
		}
	}

	topics := []models.Topic{}
	{
		repo := repository.NewRepository(&models.Topic{})
		if data.WithObjective {
			if len(data.Objective) > 0 {
				repo.Preload("Objectives", "id IN ?", data.Objective)
			} else {
				repo.Preload("Objectives")
			}
		}
		if data.WithLesson {
			repo.Preload("Objectives.Lessons")
		}

		if err := repo.FindMany(&topics, "subject_id IN ? order by subject_id asc", allowedIds); !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
	}

	progress := []models.UserProgress{}
	database.UseDB("app").Find(&progress, "user_id = ?", user.Id)

	useFilterEmptyObjective := func(arr *[]models.UserObjectiveProgress, o models.Objective) {
		if data.FilterEmptyObjective {
			if len(o.Lessons) > 0 {
				*arr = append(*arr, models.UserObjectiveProgress{
					Objective: o,
					Progress:  models.FindObjectiveProgressFromList(progress, o.Id)},
				)
			}
		} else {
			*arr = append(*arr, models.UserObjectiveProgress{
				Objective: o,
				Progress:  models.FindObjectiveProgressFromList(progress, o.Id)},
			)
		}
	}

	useFilterEmptyTopic := func(arr *[]models.UserTopicProgress, t models.Topic, pr []models.UserObjectiveProgress) {
		calc := func() (score uint) {
			score = 0
			for _, p := range pr {
				score += p.Progress
			}
			if len(pr) > 0 {
				score = uint(math.Round(float64(score) / float64(len(pr))))
			}
			return
		}
		if data.FilterEmptyObjective {
			if len(pr) > 0 {
				*arr = append(*arr, models.UserTopicProgress{
					Id:         t.Id,
					SubjectId:  t.SubjectId,
					Title:      t.Title,
					Details:    t.Details,
					Objectives: pr,
					Progress:   calc()},
				)
			}
		} else {
			*arr = append(*arr, models.UserTopicProgress{
				Id:         t.Id,
				SubjectId:  t.SubjectId,
				Title:      t.Title,
				Details:    t.Details,
				Objectives: pr,
				Progress:   calc()},
			)
		}
	}

	topicsWithProgress := []models.UserTopicProgress{}
	for _, t := range topics {
		objectivesWithProgress := []models.UserObjectiveProgress{}
		for _, o := range t.Objectives {
			useFilterEmptyObjective(&objectivesWithProgress, o)
		}
		useFilterEmptyTopic(&topicsWithProgress, t, objectivesWithProgress)
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   topicsWithProgress,
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++

//++++++++++++++++++++++++++++++++++++++++++++++++
func StudyPodcasts(ctx iris.Context) {
	type topicForm struct {
		Subject              []int
		Objective            []int
		WithObjective        bool
		FilterEmptyTopic     bool
		FilterEmptyObjective bool
	}

	data := &topicForm{}
	ctx.ReadJSON(data)
	if len(data.Subject) > 0 {
		data.Subject = list.Unique(data.Subject).([]int)
	}
	if len(data.Objective) > 0 {
		data.Objective = list.Unique(data.Objective).([]int)
	}

	user, _ := getUser(ctx)
	session := settings.Get("exam.session", time.Now().Year())
	type ID struct {
		SubjectId     int
		PaymentStatus models.PaymentStatus
	}
	ids := []ID{}
	if err := database.UseDB("app").
		Table("user_subjects as us").
		Select("us.subject_id").
		Joins("LEFT JOIN user_exams as ue on ue.id = us.user_exam_id").
		// Where("us.user_id = ? AND ue.session = ? AND ue.payment_status = ?", user.Id, session, models.Completed).
		Where("us.user_id = ? AND ue.session = ?", user.Id, session).
		Find(&ids).Error; !logger.HandleError(err) {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	allowedIds := []int{}
	{
		for _, i := range ids {
			allowedIds = append(allowedIds, i.SubjectId)
		}

		if len(data.Subject) > 0 {
			tmp := allowedIds
			allowedIds = []int{}
			for _, id := range data.Subject {
				if list.Contains(tmp, id) {
					allowedIds = append(allowedIds, id)
				}
			}
		}
	}

	topics := []models.Topic{}
	repo := repository.NewRepository(&models.Topic{})
	{
		if data.WithObjective {
			if len(data.Objective) > 0 {
				repo.Preload("Objectives", "id IN ?", data.Objective)
			} else {
				repo.Preload("Objectives")
			}
		}

		if err := repo.FindMany(&topics, "subject_id IN ? order by subject_id asc", allowedIds); !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
	}

	podcastsTopics := []models.PodcastTopic{}
	podcasts := []models.Podcast{}
	{
		pids := []int{}
		for _, t := range topics {
			podcastObjectives := []models.PodcastObjective{}
			for _, o := range t.Objectives {
				pids = append(pids, int(o.Id))
				podcastObjectives = append(podcastObjectives, models.PodcastObjective{Objective: o})
			}
			podcastsTopics = append(podcastsTopics, models.PodcastTopic{Topic: t, Objectives: podcastObjectives})
		}
		if err := database.UseDB("app").Find(&podcasts, "objective_id IN ?", pids).Error; !logger.HandleError(err) {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
	}

	progress := []models.UserProgress{}
	database.UseDB("app").Find(&progress, "user_id = ?", user.Id)

	useFilterEmptyObjective := func(arr *[]models.UserPodcastObjectiveProgress, o models.PodcastObjective) {
		if data.FilterEmptyObjective {
			if len(o.Podcasts) > 0 {
				*arr = append(*arr, models.UserPodcastObjectiveProgress{
					PodcastObjective: o,
					Progress:         models.FindObjectiveProgressFromList(progress, o.Id),
				})
			}
		} else {
			*arr = append(*arr, models.UserPodcastObjectiveProgress{
				PodcastObjective: o,
				Progress:         models.FindObjectiveProgressFromList(progress, o.Id),
			})
		}
	}

	useFilterEmptyTopic := func(arr *[]models.UserPodcastTopicProgress, t models.PodcastTopic, pr []models.UserPodcastObjectiveProgress) {
		calc := func() (score uint) {
			score = 0
			for _, p := range pr {
				score += p.Progress
			}
			if len(pr) > 0 {
				score = uint(math.Round(float64(score) / float64(len(pr))))
			}
			return
		}
		if data.FilterEmptyObjective {
			if len(pr) > 0 {
				*arr = append(*arr, models.UserPodcastTopicProgress{
					Id:         t.Id,
					SubjectId:  t.SubjectId,
					Title:      t.Title,
					Details:    t.Details,
					Progress:   calc(),
					Objectives: pr,
				})
			}
		} else {
			*arr = append(*arr, models.UserPodcastTopicProgress{
				Id:         t.Id,
				SubjectId:  t.SubjectId,
				Title:      t.Title,
				Details:    t.Details,
				Progress:   calc(),
				Objectives: pr,
			})
		}
	}

	__pt__ := []models.UserPodcastTopicProgress{}
	for _, t := range podcastsTopics {
		__po__ := []models.UserPodcastObjectiveProgress{}
		for _, o := range t.Objectives {
			useFilterEmptyObjective(&__po__, *o.FilterPodcast(podcasts))
		}
		useFilterEmptyTopic(&__pt__, t, __po__)
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   __pt__,
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++
func QuickQuiz(ctx iris.Context) {
	type quizForm struct {
		Id string
	}

	form := &quizForm{}
	ctx.ReadQuery(form)
	questions := []models.Question{}
	database.UseDB("core").Table("objective_questions as oq").
		Select("q.*").Joins("LEFT JOIN questions as q ON oq.id = q.id").
		Find(&questions, "objective_id = ?", form.Id)
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   questions,
	})
}

func QuickQuizScore(ctx iris.Context) {
	type quizForm struct {
		Id    uint
		Score uint
	}

	form := &quizForm{}
	ctx.ReadJSON(form)
	user, _ := getUser(ctx)

	objective := &models.Objective{}
	if err := database.UseDB("core").First(objective, "id = ?", form.Id).Error; err != nil {
		ctx.JSON(apiResponse{
			"status":  "failed",
			"message": "objective not found",
		})
		return
	}
	progress := &models.UserProgress{
		UserId:      user.Id,
		ObjectiveId: form.Id,
		SubjectId:   uint(objective.SubjectId),
		Score:       form.Score,
	}
	if err := database.UseDB("app").Save(progress).Error; err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	ctx.JSON(apiResponse{
		"status": "success",
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++
