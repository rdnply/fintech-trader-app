package handlers

import (
	"cw1/internal/postgres"
	"cw1/pkg/log/logger"
	"github.com/go-chi/chi"
	"net/http"
)

type Handler struct {
	logger         logger.Logger
	userStorage    *postgres.UserStorage
	sessionStorage *postgres.SessionStorage
	robotStorage   *postgres.RobotStorage
}

func NewHandler(logger logger.Logger, ut *postgres.UserStorage, st *postgres.SessionStorage, rt *postgres.RobotStorage) *Handler {
	return &Handler{
		logger:         logger,
		userStorage:    ut,
		sessionStorage: st,
		robotStorage:   rt,
	}
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
	})

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

	clientError, ok := err.(ClientError)
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
