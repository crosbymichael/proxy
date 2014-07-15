package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/crosbymichael/proxy"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
)

type Server interface {
	io.Closer
	http.Handler
}

type server struct {
	sync.Mutex

	r        *mux.Router
	backends map[string]proxy.Proxy
	logger   *logrus.Logger
	docker   *dockerclient.DockerClient
}

func New(logger *logrus.Logger, docker *dockerclient.DockerClient) Server {
	r := mux.NewRouter()

	s := &server{
		r:        r,
		logger:   logger,
		backends: make(map[string]proxy.Proxy),
		docker:   docker,
	}

	r.HandleFunc("/", s.listBackends).Methods("GET")
	r.HandleFunc("/{id:.*}", s.getBackend).Methods("GET")
	r.HandleFunc("/{id:.*}", s.addBackend).Methods("POST")
	r.HandleFunc("/{id:.*}", s.deleteBackend).Methods("DELETE")

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

func (s *server) Close() error {
	s.Lock()
	defer s.Unlock()

	var err error
	for id, p := range s.backends {
		if nerr := p.Close(); nerr != nil {
			s.logger.WithFields(logrus.Fields{
				"error": err,
				"id":    id,
			}).Error("closing backend proxy")

			err = nerr
		}
	}

	return err
}

func (s *server) listBackends(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("listing backends")

	out := []*proxy.Backend{}

	s.Lock()
	for _, p := range s.backends {
		out = append(out, p.Backend())
	}
	s.Unlock()

	s.marshal(w, out)
}

func (s *server) getBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	s.logger.WithField("id", id).Debug("getting backend")

	s.Lock()
	p, exists := s.backends[id]
	s.Unlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	s.marshal(w, p.Backend())
}

func (s *server) addBackend(w http.ResponseWriter, r *http.Request) {
	var (
		backend *proxy.Backend
		id      = mux.Vars(r)["id"]
	)

	s.logger.WithField("id", id).Debug("adding new backend")

	s.Lock()
	_, exists := s.backends[id]
	s.Unlock()

	if exists {
		http.Error(w, fmt.Sprintf("%s already exists", id), http.StatusConflict)

		return
	}

	if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
		s.logger.WithField("error", err).Error("decoding backend json")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
	backend.Name = id

	s.Lock()
	defer s.Unlock()

	proxy, err := proxy.New(backend, s.docker)
	if err != nil {
		s.logger.WithField("error", err).Error("creating new proxy")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	s.backends[id] = proxy

	if err := proxy.Start(); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err,
			"id":    id,
		}).Error("starting new proxy")

		delete(s.backends, id)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *server) deleteBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	s.logger.WithField("id", id).Info("deleting backend")

	s.Lock()
	p, exists := s.backends[id]
	s.Unlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if err := p.Close(); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err,
			"id":    id,
		}).Error("closing backend proxy")

		http.Error(w, err.Error(), http.StatusInternalServerError)

		// don't return here
	}

	s.Lock()
	delete(s.backends, id)
	s.Unlock()
}

func (s *server) marshal(w http.ResponseWriter, v interface{}) {
	w.Header().Set("content-type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.WithField("error", err).Error("marshal json")
	}
}
