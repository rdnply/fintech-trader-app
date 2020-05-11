package handlers

import (
	"cw1/cmd/auth-api/handlers/websocket"
	"cw1/cmd/auth-api/httperror"
	"cw1/internal/format"
	"cw1/internal/postgres"
	"cw1/pkg/log/logger"
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

type Handler struct {
	logger         logger.Logger
	userStorage    *postgres.UserStorage
	sessionStorage *postgres.SessionStorage
	robotStorage   *postgres.RobotStorage
	hub            *websocket.Hub
	tmplts         map[string]*template.Template
}

func NewHandler(logger logger.Logger, ut *postgres.UserStorage, st *postgres.SessionStorage, rt *postgres.RobotStorage, hb *websocket.Hub) (*Handler, error) {
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
		"printInt":      format.PrintNullInt64,
		"printFloat":    format.PrintNullFloat64,
		"printStr":      format.PrintNullString,
		"printTime":     format.PrintNullTime,
		"joinNullInt":   format.JoinNullInt,
	}

	tmplts := make(map[string]*template.Template)

	var err error

	tmplts["index"], err = template.New("index").Funcs(funcMap).ParseFiles(
		"C:\\Users\\rodion\\go\\src\\cw1\\internal\\templates\\base.html",
		"C:\\Users\\rodion\\go\\src\\cw1\\internal\\templates\\index.html")
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse index html template")
	}

	return tmplts, nil
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/signup", rootHandler{h.signUp, h.logger}.ServeHTTP)
		r.Post("/signin", rootHandler{h.signIn, h.logger}.ServeHTTP)
		r.Put("/users/{id}", rootHandler{h.updateUser, h.logger}.ServeHTTP)
		r.Get("/users/{id}", rootHandler{h.getUser, h.logger}.ServeHTTP)
		r.Get("/users/{id}/robots", rootHandler{h.getUserRobots, h.logger}.ServeHTTP)

		r.Post("/robot", rootHandler{h.createRobot, h.logger}.ServeHTTP)
		r.Delete("/robot/{id}", rootHandler{h.deleteRobot, h.logger}.ServeHTTP)
		r.Get("/robots", rootHandler{h.getRobots, h.logger}.ServeHTTP)
		r.Put("/robot/{id}/favourite", rootHandler{h.makeFavourite, h.logger}.ServeHTTP)
		r.Put("/robot/{id}/activate", rootHandler{h.activate, h.logger}.ServeHTTP)
		r.Put("/robot/{id}/deactivate", rootHandler{h.deactivate, h.logger}.ServeHTTP)
		r.Get("/robot/{id}", rootHandler{h.getRobot, h.logger}.ServeHTTP)
		r.Put("/robot/{id}", rootHandler{h.updateRobot, h.logger}.ServeHTTP)
	})
	r.HandleFunc("/ws", rootHandler{func(w http.ResponseWriter, r *http.Request) error {
		return websocket.ServeWS(h.hub, w, r)
	}, h.logger}.ServeHTTP)

	return r
}

type rootHandler struct {
	H      func(http.ResponseWriter, *http.Request) error
	logger logger.Logger
}

func (fn rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn.H(w, r)
	if err == nil {
		return
	}

	clientError, ok := err.(httperror.ClientError)
	if !ok {
		fn.logger.Errorf("Can't cast error to Client's error: %v", clientError)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	fn.logger.Errorf(clientError.Error())

	body, err := clientError.ResponseBody()
	if err != nil {
		fn.logger.Errorf("Can't get info about error because of : %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	status, headers := clientError.ResponseHeaders()
	for k, v := range headers {
		if body == nil && v == "application/json" {
			continue
		}

		w.Header().Set(k, v)
	}

	w.WriteHeader(status)

	c, err := w.Write(body)
	if err != nil {
		fn.logger.Errorf("Can't write json data in respond, code: %v, error: %v", c, err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
