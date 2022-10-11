package controllers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/list"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/repository"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

//++++++++++++++++++++++++++++++++++++++++++++++++
func StudySubjects(ctx iris.Context) {
	type (
		subjectForm struct {
			Exam          []string
			WithObjective bool
		}
	)
	var response map[string][]models.Subject = make(map[string][]models.Subject)

	data := &subjectForm{}
	ctx.ReadJSON(data)
	user, _ := getUser(ctx)
	session := settings.Get("examSession", time.Now().Year())
	examSubject := map[query][]models.UserSubject{}
	{
		q := []query{}
		if err := database.UseDB("app").Table("user_exams as ue").
			Select("ue.id, ue.exam_id, ue.user_id, ue.session, e.name, e.status").
			Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
			Where("ue.session = ? AND e.status = 1 AND ue.user_id = ?", session, user.Id).
			Scan(&q).Error; err != nil {
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
	if data.WithObjective {
		repo.Preload("Objectives")
	}
	for e := range examSubject {
		subjects := []models.Subject{}
		if err := repo.FindMany(&subjects, "id IN ?", ids[e.Name]); err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(internalServerError)
			return
		}
		response[e.Name] = append(response[e.Name], subjects...)
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   response,
	})
}

func StudyTopics(ctx iris.Context) {
	type topicForm struct {
		Subject       []int
		Objective     []int
		WithObjective bool
		WithLesson    bool
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
	session := settings.Get("examSession", time.Now().Year())
	type ID struct {
		SubjectId     int
		PaymentStatus models.PaymentStatus
	}
	ids := []ID{}
	if err := database.UseDB("app").
		Table("user_subjects as us").
		Select("us.subject_id").
		Joins("LEFT JOIN user_exams as ue on ue.id = us.user_exam_id").
		Where("us.user_id = ? AND ue.session = ? AND ue.payment_status = ?", user.Id, session, models.Completed).
		Find(&ids).Error; err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	allowedIds := []int{}
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

	topics := []models.Topic{}
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

	if err := repo.FindMany(&topics, "subject_id IN ? order by subject_id asc", allowedIds); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	if len(data.Objective) > 0 {
		tmp := topics
		topics = []models.Topic{}
		for _, t := range tmp {
			tmpObj := t.Objectives
			t.Objectives = []models.TopicObjective{}
			if len(data.Objective) > 0 {
				for _, o := range tmpObj {
					if list.Contains(data.Objective, int(o.Id)) {
						t.Objectives = append(t.Objectives, o)
					}
				}
				topics = append(topics, t)
			} else {
				topics = append(topics, t)
			}
		}
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   topics,
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++

//++++++++++++++++++++++++++++++++++++++++++++++++
func StudyPodcasts(ctx iris.Context) {
	type topicForm struct {
		Subject       []int
		Objective     []int
		WithObjective bool
		WithLesson    bool
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
	session := settings.Get("examSession", time.Now().Year())
	type ID struct {
		SubjectId     int
		PaymentStatus models.PaymentStatus
	}
	ids := []ID{}
	if err := database.UseDB("app").
		Table("user_subjects as us").
		Select("us.subject_id").
		Joins("LEFT JOIN user_exams as ue on ue.id = us.user_exam_id").
		Where("us.user_id = ? AND ue.session = ? AND ue.payment_status = ?", user.Id, session, models.Completed).
		Find(&ids).Error; err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	allowedIds := []int{}
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

	topics := []models.Topic{}
	repo := repository.NewRepository(&models.Topic{})
	if data.WithObjective {
		if len(data.Objective) > 0 {
			repo.Preload("Objectives", "id IN ?", data.Objective)
		} else {
			repo.Preload("Objectives")
		}
	}

	if err := repo.FindMany(&topics, "subject_id in ? order by subject_id asc", allowedIds); err != nil {
		ctx.StatusCode(500)
		ctx.JSON(internalServerError)
		return
	}

	podcastsTopic := []models.PodcastTopic{}
	for _, t := range topics {
		podcastObjective := []models.PodcastObjective{}
		for _, o := range t.Objectives {
			if len(data.Objective) > 0 {
				if list.Contains(data.Objective, int(o.Id)) {
					po := models.PodcastObjective{
						Id:          o.Id,
						TopicId:     o.TopicId,
						ObjectiveId: o.ObjectiveId,
						Title:       o.Title,
						Details:     o.Details,
					}
					po.Podcasts = po.GetPodcasts()
					podcastObjective = append(podcastObjective, po)
				}
			} else {
				po := models.PodcastObjective{
					Id:          o.Id,
					TopicId:     o.TopicId,
					ObjectiveId: o.ObjectiveId,
					Title:       o.Title,
					Details:     o.Details,
				}
				po.Podcasts = po.GetPodcasts()
				podcastObjective = append(podcastObjective, po)
			}
		}
		podcastsTopic = append(podcastsTopic, models.PodcastTopic{
			Id:         t.Id,
			SubjectId:  t.SubjectId,
			Title:      t.Title,
			Details:    t.Details,
			Objectives: podcastObjective,
		})
	}

	ctx.JSON(apiResponse{
		"status": "success",
		"data":   podcastsTopic,
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++
func QuickQuiz(ctx iris.Context) {
	type quizForm struct {
		Subject   []uint
		Objective []uint
	}

	form := &quizForm{}
	ctx.ReadJSON(form)
	user, _ := getUser(ctx)
	session := settings.Get("examSession", time.Now().Year())

	q := []queryWithPayment{}
	if err := database.UseDB("app").Table("user_exams as ue").
		Select("ue.id, ue.exam_id, ue.session, ue.user_id, e.name, e.status, ue.payment_status").
		Joins("LEFT JOIN exams as e on ue.exam_id = e.id").
		Where("ue.session = ? AND e.status = 1 AND ue.payment_status = ? AND ue.user_id = ?", session, models.Completed, user.Id).
		Scan(&q).Error; err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}

	allowedSubjectIds, err := func(qp []queryWithPayment) (ids []uint, err error) {
		for _, q := range qp {
			if q.Id != uuid.Nil {
				userSubjects := []models.UserSubject{}
				if err = repository.NewRepository(&models.UserSubject{}).
					FindMany(&userSubjects, "user_exam_id = ?", q.Id); err != nil {
					return
				}
				for _, us := range userSubjects {
					ids = append(ids, us.SubjectId)
				}
			}
		}
		return
	}(q)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(internalServerError)
		return
	}
	// Filter subjects to registered only
	{
		ids := []uint{}
		for _, id := range form.Subject {
			if list.Contains(allowedSubjectIds, id) {
				ids = append(ids, id)
			}
		}
		form.Subject = ids
	}

	questions := []models.Question{}
	if len(allowedSubjectIds) == 0 {
		for _, id := range form.Subject {
			_questions := []models.Question{}
			if err := database.UseDB("core").Find(&_questions, "subject_id = ? LIMIT 4", id).Error; err != nil {
				ctx.StatusCode(500)
				ctx.JSON(internalServerError)
				return
			}
			questions = append(questions, _questions...)
		}
		ctx.JSON(apiResponse{
			"status": "success",
			"data":   models.SortBySubject(questions),
		})
		return
	}

	quickQuizQuestions := settings.Get("quickQuizQuestions", 60).(int)
	rand.Seed(time.Now().Unix() * rand.Int63())
	if hasObjectiveFilter := len(form.Objective) > 0; !hasObjectiveFilter {
		subs := []models.Subject{}
		if err := repository.NewRepository(&models.Subject{}).
			Preload("Objectives.Questions").
			FindMany(&subs, "id IN ?", form.Subject); err != nil {
			ctx.StatusCode(500)
			ctx.JSON(internalServerError)
			return
		}
		for _, s := range subs {
			ids := []uint{}
			for _, o := range s.Objectives {
				for _, q := range o.Questions {
					ids = append(ids, uint(q.QuestionId))
				}
			}
			if len(ids) == 0 {
				ids = append(ids, 0)
			}

			q := []interface{}{}
			if l := len(ids); l >= (quickQuizQuestions) {
				r := []uint{}
				for i := len(ids) - 1; i != 0; i-- {
					rIndex := rand.Intn(i)
					r = append(r, ids[rIndex])
				}
				ids = r[:quickQuizQuestions]
				q = append(q, "id IN ? AND question_type_id = ? AND subject_id = ? ORDER BY RAND() LIMIT ?", ids, models.OBJECTIVE, s.Id, quickQuizQuestions)
			} else {
				q = append(q, "id IN ? AND question_type_id = ? AND subject_id = ?  ORDER BY RAND() LIMIT ?", ids, models.OBJECTIVE, s.Id, quickQuizQuestions)
			}
			_questions := []models.Question{}
			database.UseDB("core").Find(&_questions, q...)
			questions = append(questions, _questions...)
			l := len(_questions)
			if r := quickQuizQuestions - l; l < quickQuizQuestions {
				database.UseDB("core").
					Find(&_questions, "id NOT IN ? AND question_type_id = ? AND subject_id = ? ORDER BY RAND() LIMIT ?", ids, models.OBJECTIVE, s.Id, r+1)
				questions = append(questions, _questions[:quickQuizQuestions-l+1]...)
			}
		}
	} else {
		database.UseDB("core").Table("questions as q").
			Select("*").
			Joins("LEFT JOIN objective_questions oq on oq.id  = q.id").
			Joins("LEFT JOIN objectives o on o.id = oq.objective_id").
			Where("o.id IN ? AND q.subject_id IN ? AND question_type_id = ? ORDER BY RAND() LIMIT ?", form.Objective, form.Subject, models.OBJECTIVE, len(form.Subject)*quickQuizQuestions).
			Scan(&questions)
	}
	ctx.JSON(apiResponse{
		"status": "success",
		"data":   models.SortBySubject(questions),
	})
}

//++++++++++++++++++++++++++++++++++++++++++++++++
