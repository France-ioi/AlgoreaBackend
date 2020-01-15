// Package answers provides API services for task answers managing
package answers

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// Service is the mount point for services related to `answers`
type Service struct {
	service.Base
}

// swagger:model
type answerData struct {
	// required: true
	// minLength: 1
	Answer string `json:"answer" validate:"set,min=1"`
	// required: true
	// minLength: 1
	State string `json:"state" validate:"set,min=1"`
}

// SetRoutes defines the routes for this package in a route answers
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))
	router.Get("/answers", service.AppHandler(srv.getAnswers).ServeHTTP)
	router.Get("/answers/{answer_id}", service.AppHandler(srv.get).ServeHTTP)
	router.Post("/answers", service.AppHandler(srv.submit).ServeHTTP)
	router.Post("/attempts/{attempt_id}/answers", service.AppHandler(srv.save).ServeHTTP)
	router.Put("/attempts/{attempt_id}/answers/current", service.AppHandler(srv.updateCurrent).ServeHTTP)
}
