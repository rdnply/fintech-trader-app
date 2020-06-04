package handler

import (
	"cw1/cmd/socket"
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"cw1/internal/user"
	"cw1/pkg/log/logger"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

type Handler struct {
	logger         logger.Logger
	userStorage    user.Storage
	sessionStorage session.Storage
	robotStorage   robot.Storage
	hub            *socket.Hub
	tmplts         map[string]*template.Template
}

func New(logger logger.Logger, ut user.Storage, st session.Storage,
	rt robot.Storage, hb *socket.Hub) (*Handler, error) {
	t, err := parseTemplates()
	if err != nil {
		return nil, errors.Wrap(err, "can't parse templates for handler")
	}

	return &Handler{
		logger:         logger,
		userStorage:    ut,
		sessionStorage: st,
		robotStorage:   rt,
		tmplts:         t,
		hub:            hb,
	}, nil
}

func parseTemplates() (map[string]*template.Template, error) {
	funcMap := template.FuncMap{
		"printInt":    format.PrintNullInt64,
		"printFloat":  format.PrintNullFloat64,
		"printStr":    format.PrintNullString,
		"printTime":   format.PrintNullTime,
		"joinNullInt": format.JoinNullInt,
	}

	tmplts := make(map[string]*template.Template)

	var err error

	_ = os.Chdir("./internal/templates")

	pwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrapf(err, "can't get path")
	}

	tmplts["index"], err = template.New("index").Funcs(funcMap).
		ParseFiles(fmt.Sprintf("%s/base.html", pwd), fmt.Sprintf("%s/index.html", pwd))
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse index html template")
	}

	return tmplts, nil
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/signup", h.signUp)
		r.Post("/signin", h.signIn)
		r.Put("/users/{id}", h.updateUser)
		r.Get("/users/{id}", h.getUser)
		r.Get("/users/{id}/robots", h.getUserRobots)

		r.Post("/robot", h.createRobot)
		r.Delete("/robot/{id}", h.deleteRobot)
		r.Get("/robots", h.getRobots)
		r.Put("/robot/{id}/favourite", h.makeFavourite) //nolint: misspell
		r.Put("/robot/{id}/activate", h.activate)
		r.Put("/robot/{id}/deactivate", h.deactivate)
		r.Get("/robot/{id}", h.getRobot)
		r.Put("/robot/{id}", h.updateRobot)
	})

	r.HandleFunc("/ws", func(w http.ResponseWriter, rr *http.Request) {
		socket.ServeWS(h.hub, w, rr)
	})

	return r
}
