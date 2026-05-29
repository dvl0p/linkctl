package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/dvl0p/linkctl/internal/db"
)

type LinkRequest struct {
	Url             string `json:"url"`
	IntervalSeconds int64  `json:"interval_seconds"`
}

type Link struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Url             string    `json:"url"`
	IntervalSeconds int64     `json:"interval_seconds"`
}

func (s *Server) handlerCreateLink(w http.ResponseWriter,
	r *http.Request) (int, error) {

	decoder := json.NewDecoder(r.Body)
	link := LinkRequest{}
	if err := decoder.Decode(&link); err != nil {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "could not parse request")
		return code, err
	}

	linkDB, err := s.store.Queries.CreateLink(r.Context(), db.CreateLinkParams{
		Url:             link.Url,
		IntervalSeconds: link.IntervalSeconds,
	})
	if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}

	code := http.StatusCreated
	handlerReplyWithJSON(w, code, Link{
		ID:              linkDB.ID,
		CreatedAt:       linkDB.CreatedAt,
		UpdatedAt:       linkDB.UpdatedAt,
		Url:             linkDB.Url,
		IntervalSeconds: linkDB.IntervalSeconds,
	})
	return code, nil
}

func (s *Server) handlerListLinks(w http.ResponseWriter,
	r *http.Request) (int, error) {
	if r.URL.Query().Has("id") {
		return s.handlerGetLink(w, r)
	}
	if r.URL.Query().Has("url") {
		return s.handlerGetLinkFromURL(w, r)
	}

	linksDB, err := s.store.Queries.ListLinks(r.Context())
	if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}

	links := make([]Link, len(linksDB))
	for i, linkDB := range linksDB {
		links[i] = Link{
			ID:              linkDB.ID,
			CreatedAt:       linkDB.CreatedAt,
			UpdatedAt:       linkDB.UpdatedAt,
			Url:             linkDB.Url,
			IntervalSeconds: linkDB.IntervalSeconds,
		}
	}

	sortLinks(links, r.URL.Query().Get("sort"))

	code := http.StatusOK
	handlerReplyWithJSON(w, code, links)
	return code, nil
}

func (s *Server) handlerGetLink(w http.ResponseWriter,
	r *http.Request) (int, error) {

	linkIDStr := r.URL.Query().Get("id")
	if linkIDStr == "" {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid request")
		return code, errors.New("link query missing link id")
	}

	linkID, err := strconv.ParseInt(linkIDStr, 10, 64)
	if err != nil {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid link id")
		return code, errors.New("link query contains malformed link id")
	}

	linkDB, err := s.store.Queries.GetLink(r.Context(), linkID)
	if err == sql.ErrNoRows {
		code := http.StatusNotFound
		handlerReplyWithError(w, code, "no matching link id was found")
		return code, err
	}
	if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}

	code := http.StatusOK
	handlerReplyWithJSON(w, code, Link{
		ID:              linkDB.ID,
		CreatedAt:       linkDB.CreatedAt,
		UpdatedAt:       linkDB.UpdatedAt,
		Url:             linkDB.Url,
		IntervalSeconds: linkDB.IntervalSeconds,
	})
	return code, nil
}

func (s *Server) handlerGetLinkFromURL(w http.ResponseWriter,
	r *http.Request) (int, error) {

	urlStr := r.URL.Query().Get("url")
	if urlStr == "" {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid request")
		return code, errors.New("link query missing url")
	}

	linkDB, err := s.store.Queries.GetLinkFromURL(r.Context(), urlStr)
	if err == sql.ErrNoRows {
		code := http.StatusNotFound
		handlerReplyWithError(w, code, "no matching link url found")
		return code, err
	}
	if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}

	code := http.StatusOK
	handlerReplyWithJSON(w, code, Link{
		ID:              linkDB.ID,
		CreatedAt:       linkDB.CreatedAt,
		UpdatedAt:       linkDB.UpdatedAt,
		Url:             linkDB.Url,
		IntervalSeconds: linkDB.IntervalSeconds,
	})
	return code, nil
}

func (s *Server) handlerUpdateLink(w http.ResponseWriter,
	r *http.Request) (int, error) {
	if r.URL.Query().Has("id") {
		return s.handlerUpdateLinkFromID(w, r)
	}
	if r.URL.Query().Has("url") {
		return s.handlerUpdateLinkFromURL(w, r)
	}
	code := http.StatusBadRequest
	handlerReplyWithError(w, code, "invalid request")
	return code, errors.New("update request missing query specification")
}

func (s *Server) handlerUpdateLinkFromID(w http.ResponseWriter,
	r *http.Request) (int, error) {
	linkIDStr := r.URL.Query().Get("id")
	if linkIDStr == "" {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid request")
		return code, errors.New("link update request missing link id")
	}

	linkID, err := strconv.ParseInt(linkIDStr, 10, 64)
	if err != nil {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid link id")
		return code, err
	}

	type updateLinkRequest struct {
		IntervalSeconds int64 `json:"interval_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	req := updateLinkRequest{}
	if err := decoder.Decode(&req); err != nil {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "malformed update request")
		return code, err
	}

	if _, err := s.store.Queries.UpdateLink(r.Context(), db.UpdateLinkParams{
		ID:              linkID,
		IntervalSeconds: req.IntervalSeconds,
	}); err == sql.ErrNoRows {
		code := http.StatusNotFound
		handlerReplyWithError(w, code, "no matching link id was found")
		return code, err
	} else if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}

	code := http.StatusNoContent
	w.WriteHeader(code)
	return code, nil
}

func (s *Server) handlerUpdateLinkFromURL(w http.ResponseWriter,
	r *http.Request) (int, error) {

	urlStr := r.URL.Query().Get("url")
	if urlStr == "" {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid request")
		return code, errors.New("link update request missing link url")
	}

	type updateLinkRequest struct {
		IntervalSeconds int64 `json:"interval_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	req := updateLinkRequest{}
	if err := decoder.Decode(&req); err != nil {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "malformed update request")
		return code, err
	}

	if _, err := s.store.Queries.UpdateLinkFromURL(r.Context(),
		db.UpdateLinkFromURLParams{
			Url:             urlStr,
			IntervalSeconds: req.IntervalSeconds,
		}); err == sql.ErrNoRows {
		code := http.StatusNotFound
		handlerReplyWithError(w, code, "no matching link url was found")
		return code, err
	} else if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}

	code := http.StatusNoContent
	w.WriteHeader(code)
	return code, nil
}

func (s *Server) handlerDeleteLink(w http.ResponseWriter,
		r *http.Request) (int, error) {

	if r.URL.Query().Has("id") {
		return s.handlerDeleteLinkFromID(w, r)
	}
	if r.URL.Query().Has("url") {
		return s.handlerDeleteLinkFromURL(w, r)
	}
	code := http.StatusBadRequest
	handlerReplyWithError(w, code, "invalid request")
	return code, errors.New("delete request missing query specification")
}

func (s *Server) handlerDeleteLinkFromID(w http.ResponseWriter,
		r *http.Request) (int, error) {
	linkIDStr := r.URL.Query().Get("id")
	if linkIDStr == "" {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid request")
		return code, errors.New("delete request missing link id")
	}

	linkID, err := strconv.ParseInt(linkIDStr, 10, 64)
	if err != nil {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid link id")
		return code, err
	}

	if _, err := s.store.Queries.DeleteLink(r.Context(),
			linkID); err == sql.ErrNoRows {
		code := http.StatusNotFound
		handlerReplyWithError(w, code, "no matching link id was found")
		return code, err
	} else if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}
	code := http.StatusNoContent
	w.WriteHeader(code)
	return code, nil
}

func (s *Server) handlerDeleteLinkFromURL(w http.ResponseWriter,
		r *http.Request) (int, error) {

	urlStr := r.URL.Query().Get("url")
	if urlStr == "" {
		code := http.StatusBadRequest
		handlerReplyWithError(w, code, "invalid request")
		return code, errors.New("delete request missing link url")
	}

	if _, err := s.store.Queries.DeleteLinkFromURL(r.Context(),
			urlStr); err == sql.ErrNoRows {
		code := http.StatusNotFound
		handlerReplyWithError(w, code, "no matching link url was found")
		return code, err
	} else if err != nil {
		code := http.StatusInternalServerError
		handlerReplyWithError(w, code, "database error")
		return code, err
	}
	code := http.StatusNoContent
	w.WriteHeader(code)
	return code, nil
}

func sortLinks(links []Link, sortOrder string) {
	if sortOrder == "desc" {
		slices.SortFunc(links, func(a, b Link) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})
	}
}
