package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nurmuh-alhakim18/chirpy/internal/auth"
	"github.com/nurmuh-alhakim18/chirpy/internal/database"
)

type chirp struct {
	Id        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- CREATE CHIRP ---
func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respError(w, 401, "Couldn't find JWT", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secretKeyJWT)
	if err != nil {
		respError(w, 401, "Couldn't validate JWT", err)
		return
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respError(w, 500, "Couldn't decode parameters", err)
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirpCreated, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: userId,
		Body:   cleaned,
	})

	if err != nil {
		respError(w, 500, "Couldn't create chirp", err)
		return
	}

	respJSON(w, 201, chirp{
		Id:        chirpCreated.ID,
		UserId:    chirpCreated.UserID,
		Body:      chirpCreated.Body,
		CreatedAt: chirpCreated.CreatedAt,
		UpdatedAt: chirpCreated.UpdatedAt,
	})
}

func validateChirp(body string) (string, error) {
	maxChirpLength := 140
	if len(body) > maxChirpLength {
		return "", errors.New("chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleanedBody := cleanBody(body, badWords)

	return cleanedBody, nil
}

func cleanBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		lowered := strings.ToLower(word)
		if _, ok := badWords[lowered]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

// --- GET ALL CHIRPS ---
func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirpsDB, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respError(w, 500, "Couldn't get chirps", err)
		return
	}

	var authorId uuid.UUID
	authorIdString := r.URL.Query().Get("author_id")
	if authorIdString != "" {
		authorId, err = uuid.Parse(authorIdString)
		if err != nil {
			respError(w, 500, "Couldn't parse author id", err)
			return
		}
	}

	sortDir := "asc"
	sortQuery := r.URL.Query().Get("sort")
	if sortQuery == "desc" {
		sortDir = "desc"
	}

	var chirps []chirp
	for _, chirpDB := range chirpsDB {
		if authorId != uuid.Nil && chirpDB.UserID != authorId {
			continue
		}

		chirps = append(chirps, chirp{
			Id:        chirpDB.ID,
			UserId:    chirpDB.UserID,
			Body:      chirpDB.Body,
			CreatedAt: chirpDB.CreatedAt,
			UpdatedAt: chirpDB.UpdatedAt,
		})
	}

	if sortDir == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}

	respJSON(w, 200, chirps)
}

// --- GET CHIRP BY ID ---
func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpIdString := r.PathValue("chirpId")
	chirpID, err := uuid.Parse(chirpIdString)
	if err != nil {
		respError(w, 400, "Invalid chirp ID", err)
		return
	}

	chirpDB, err := cfg.db.GetChirpById(r.Context(), chirpID)
	if err != nil {
		respError(w, 404, "Couldn't get chirp", err)
	}

	respJSON(w, 200, chirp{
		Id:        chirpDB.ID,
		UserId:    chirpDB.UserID,
		Body:      chirpDB.Body,
		CreatedAt: chirpDB.CreatedAt,
		UpdatedAt: chirpDB.UpdatedAt,
	})
}

// --- DELETE CHIRP ---
func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpIdString := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(chirpIdString)
	if err != nil {
		respError(w, 400, "Invalid chirp ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respError(w, 401, "Couldn't find JWT", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secretKeyJWT)
	if err != nil {
		respError(w, 401, "Couldn't validate JWT", err)
		return
	}

	chirpDB, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respError(w, 404, "Couldn't get chirp", err)
		return
	}

	if chirpDB.UserID != userId {
		respError(w, 403, "Couldn't delete chirp", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		respError(w, 500, "Couldn't delete chirp", err)
	}

	w.WriteHeader(204)
}
