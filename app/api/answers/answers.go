// Package answers provides API services for task answers managing
package answers

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `answers`
type Service struct {
	*service.Base
}

// swagger:model
type answerData struct {
	// required: true
	Answer string `json:"answer" validate:"set"`
	// required: true
	State string `json:"state" validate:"set"`
}

// SetRoutes defines the routes for this package in a route answers
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Base))
	router.Get("/items/{item_id}/answers", service.AppHandler(srv.listAnswers).ServeHTTP)
	router.Get("/answers/{answer_id}", service.AppHandler(srv.getAnswer).ServeHTTP)
	router.Post("/answers", service.AppHandler(srv.submit).ServeHTTP)
	router.Post("/answers/{answer_id}/generate-task-token", service.AppHandler(srv.generateTaskToken).ServeHTTP)

	routerWithParticipant := router.With(service.ParticipantMiddleware(srv.Base))
	routerWithParticipant.Get("/items/{item_id}/current-answer", service.AppHandler(srv.getCurrentAnswer).ServeHTTP)
	routerWithParticipant.Post("/items/{item_id}/attempts/{attempt_id}/answers", service.AppHandler(srv.answerCreate).ServeHTTP)
	routerWithParticipant.Put("/items/{item_id}/attempts/{attempt_id}/answers/current", service.AppHandler(srv.updateCurrentAnswer).ServeHTTP)
}

func withGradings(answersQuery *database.DB) *database.DB {
	return answersQuery.
		Joins("LEFT JOIN gradings ON gradings.answer_id = answers.id").
		Select(`answers.id, answers.author_id, answers.item_id, answers.attempt_id, answers.participant_id,
			answers.type, answers.state, answers.answer, answers.created_at, gradings.score,
			gradings.graded_at`)
}
