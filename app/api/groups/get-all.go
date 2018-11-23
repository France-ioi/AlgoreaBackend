package groups

import (
  "net/http"

  s "github.com/France-ioi/AlgoreaBackend/app/service"
  "github.com/go-chi/render"
)

// GroupResponseRow is the structure of a row for the service response
type GroupResponseRow struct {
  ID   int    `json:"id" db:"ID"`
  Name string `json:"name" db:"sName"`
}

func (srv *Service) getAll(w http.ResponseWriter, r *http.Request) s.APIError {
  var err error

  groups := []GroupResponseRow{}
  if err = srv.Store.Groups.GetAll(&groups); err != nil {
    return s.ErrUnexpected(err)
  }
  render.Respond(w, r, groups)
  return s.NoError
}
