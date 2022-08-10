package models

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/dbmodel"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"gorm.io/gorm"
)

type (
	Question struct {
		Id                 uint      `sql:"primary_key;" json:"id"`
		SubjectId          uint      `json:"subject_id"`
		SourceId           uint      `json:"source_id"`
		QuestionTypeId     uint      `json:"question_type_id"`
		Question           string    `json:"question"`
		QuestionDetails    string    `json:"question_details"`
		QuestionImage      string    `json:"question_image"`
		Option1            string    `json:"option_1"`
		Option2            string    `json:"option_2"`
		Option3            string    `json:"option_3"`
		Option4            string    `json:"option_4"`
		ShortAnswer        string    `json:"short_answer"`
		FullAnswer         string    `json:"full_answer"`
		AnswerImage        string    `json:"answer_image"`
		AnswerDetails      string    `json:"answer_details"`
		QuestionYear       uint      `json:"question_year"`
		QuestionYearNumber uint      `json:"question_year_number"`
		CreatedAt          time.Time `json:"-"`
		UpdatedAt          time.Time `json:"-"`
	}
)

const (
	OBJECTIVE = 1
	THEORY    = 2
	PRACTICAL = 3
)

func (u *Question) ID() interface{} {
	return u.Id
}

func (u *Question) Tag() string {
	return "questions"
}

func (u *Question) Database() *gorm.DB {
	return database.UseDB("core")
}

func (u *Question) Migrate() dbmodel.Migration {
	return dbmodel.NewMigration(u)
}

func RandomizeQuestions(arr []Question) (r []Question) {
	rand.Seed(time.Now().Unix() * rand.Int63())
	r = make([]Question, 0)
	for i := len(arr) - 1; i > 0; i-- {
		rIndex := rand.Intn(i)
		r = append(r, arr[rIndex])
	}
	return r
}

type SortedQuestion struct {
	Subject
	Questions []Question `json:"questions"`
}

func SortBySubject(arr []Question) (r []SortedQuestion) {
	r = make([]SortedQuestion, 0)
	_r := map[uint]SortedQuestion{}
	for _, q := range arr {
		s, ok := _r[q.SubjectId]
		if !ok {
			if err := database.UseDB("core").Find(&s.Subject, "id = ?", q.SubjectId).Error; err != nil {
				fmt.Println("skip")
				continue
			}
			_r[s.Id] = SortedQuestion{Subject: s.Subject}
		}
		qs := _r[s.Id]
		qs.Questions = append(_r[s.Id].Questions, q)
		_r[s.Id] = qs
	}
	for _, v := range _r {
		r = append(r, SortedQuestion{v.Subject, RandomizeQuestions(v.Questions)})
	}
	return
}
