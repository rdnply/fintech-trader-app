package handler

import (
	"cw1/cmd/socket"
	"cw1/internal/format"
	"cw1/internal/robot"
	"cw1/internal/session"
	"cw1/internal/user"
	"cw1/pkg/log/logger"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"html/template"
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
		r.Post("/signup", h.signUp)
		r.Post("/signin", h.signIn)
		r.Put("/users/{id}", h.updateUser)
		r.Get("/users/{id}", h.getUser)
		r.Get("/users/{id}/robots", h.getUserRobots)

		r.Post("/robot", h.createRobot)
		r.Delete("/robot/{id}", h.deleteRobot)
		r.Get("/robots", h.getRobots)
		r.Put("/robot/{id}/favourite", h.makeFavourite)
		r.Put("/robot/{id}/activate", h.activate)
		r.Put("/robot/{id}/deactivate", h.deactivate)
		r.Get("/robot/{id}", h.getRobot)
		r.Put("/robot/{id}", h.updateRobot)
	})
	return r
}


//	})
//	rr.HandleFunc("/ws", rootHandler{func(w http.ResponseWriter, rr *http.Request) error {
//		return socket.ServeWS(h.hub, w, rr)
//	}, h.logger}.ServeHTTP)
//
//	return rr
//}

//type rootHandler struct {
//	H      func(http.ResponseWriter, *http.Request) error
//	logger logger.Logger
//}
//
//func (fn rootHandler) ServeHTTP(w http.ResponseWriter, rr *http.Request) {
//	err := fn.H(w, rr)
//	if err == nil {
//		return
//	}
//
//	clientError, ok := err.(render.ClientError)
//	if !ok {
//		fn.logger.Errorf("Can't cast error to Client's error: %v", clientError)
//		w.WriteHeader(http.StatusInternalServerError)
//
//		return
//	}
//
//	fn.logger.Errorf(clientError.Error())
//
//	body, err := clientError.ResponseBody()
//	if err != nil {
//		fn.logger.Errorf("Can't get info about error because of : %v", err)
//		w.WriteHeader(http.StatusInternalServerError)
//
//		return
//	}
//
//	status, headers := clientError.ResponseHeaders()
//	for k, v := range headers {
//		if body == nil && v == "application/json" {
//			continue
//		}
//
//		w.Header().Set(k, v)
//	}
//
//	w.WriteHeader(status)
//
//	c, err := w.Write(body)
//	if err != nil {
//		fn.logger.Errorf("Can't write json data in respond, code: %v, error: %v", c, err)
//		w.WriteHeader(http.StatusInternalServerError)
//
//		return
//	}
//}
