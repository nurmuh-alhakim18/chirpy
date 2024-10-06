package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nurmuh-alhakim18/chirpy/internal/auth"
	"github.com/nurmuh-alhakim18/chirpy/internal/database"
)

type user struct {
	Id          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Password    string    `json:"password,omitempty"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respError(w, 500, "Couldn't decode parameters", err)
		return
	}

	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respError(w, 500, "Couldn't hash password", err)
		return
	}

	userCreated, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	})

	if err != nil {
		respError(w, 500, "Couldn't create user", err)
		return
	}

	respJSON(w, 201, user{
		Id:          userCreated.ID,
		Email:       userCreated.Email,
		IsChirpyRed: userCreated.IsChirpyRed,
		CreatedAt:   userCreated.CreatedAt,
		UpdatedAt:   userCreated.UpdatedAt,
	})
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		user
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respError(w, 500, "Couldn't decode parameters", err)
		return
	}

	userDB, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respError(w, 401, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, userDB.HashedPassword)
	if err != nil {
		respError(w, 401, "Incorrect email or password", err)
		return
	}

	token, err := auth.MakeJWT(userDB.ID, cfg.secretKeyJWT, time.Hour)
	if err != nil {
		respError(w, 500, "Couldn't create access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respError(w, 500, "Couldn't create refresh token", err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userDB.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})

	if err != nil {
		respError(w, 500, "Couldn't save refresh token", err)
	}

	respJSON(w, 200, response{
		user: user{
			Id:          userDB.ID,
			Email:       userDB.Email,
			IsChirpyRed: userDB.IsChirpyRed,
			CreatedAt:   userDB.CreatedAt,
			UpdatedAt:   userDB.UpdatedAt,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respError(w, 500, "Couldn't decode parameters", err)
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

	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respError(w, 500, "Couldn't hash password", err)
		return
	}

	userUpdated, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
		ID:             userId,
	})

	if err != nil {
		respError(w, 500, "Couldn't update user data", err)
	}

	respJSON(w, 200, user{
		Id:          userUpdated.ID,
		Email:       userUpdated.Email,
		IsChirpyRed: userUpdated.IsChirpyRed,
		CreatedAt:   userUpdated.CreatedAt,
		UpdatedAt:   userUpdated.UpdatedAt,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respError(w, 400, "Couldn't find token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respError(w, 401, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.secretKeyJWT, time.Hour)
	if err != nil {
		respError(w, 401, "Couldn't validate token", err)
		return
	}

	respJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respError(w, 400, "Couldn't find token", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpgradeChirpyRed(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respError(w, 500, "Couldn't decode parameters", err)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respError(w, 401, "Couldn't find api key", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respError(w, 401, "api key is invalid", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	err = cfg.db.UpgradeChirpyRed(r.Context(), params.Data.UserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respError(w, 404, "Couldn't find user", err)
		}

		respError(w, 500, "Couldn't update user", err)
		return
	}

	w.WriteHeader(204)
}
