// Package answers provides API services for task answers managing.
package answers

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `answers`.
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

// SetRoutes defines the routes for this package in a route answers.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Post("/answers", service.AppHandler(srv.submit).ServeHTTP)

	routerWithAuth := router.With(auth.UserMiddleware(srv.Base))
	routerWithAuth.Get("/items/{item_id}/answers", service.AppHandler(srv.listAnswers).ServeHTTP)
	routerWithAuth.Get("/items/{item_id}/best-answer", service.AppHandler(srv.getBestAnswer).ServeHTTP)
	routerWithAuth.Get("/answers/{answer_id}", service.AppHandler(srv.getAnswer).ServeHTTP)
	routerWithAuth.Post("/answers/{answer_id}/generate-task-token", service.AppHandler(srv.generateTaskToken).ServeHTTP)

	routerWithParticipant := routerWithAuth.With(service.ParticipantMiddleware(srv.Base))
	routerWithParticipant.Get("/items/{item_id}/current-answer", service.AppHandler(srv.getCurrentAnswer).ServeHTTP)
	routerWithParticipant.Post("/items/{item_id}/attempts/{attempt_id}/answers", service.AppHandler(srv.answerCreate).ServeHTTP)
	routerWithParticipant.Put("/items/{item_id}/attempts/{attempt_id}/answers/current", service.AppHandler(srv.updateCurrentAnswer).ServeHTTP)
}
