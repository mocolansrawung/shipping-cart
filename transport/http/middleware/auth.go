package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/evermos/boilerplate-go/configs"
	"github.com/evermos/boilerplate-go/infras"
	"github.com/evermos/boilerplate-go/shared"
	"github.com/evermos/boilerplate-go/shared/failure"
	"github.com/evermos/boilerplate-go/shared/oauth"
	"github.com/evermos/boilerplate-go/transport/http/response"
)

type Authentication struct {
	db     *infras.MySQLConn
	config *configs.Config
}

type ValidateAuthResponse struct {
	Data shared.Claims `json:"data"`
}

const (
	HeaderAuthorization = "Authorization"
)

func ProvideAuthentication(db *infras.MySQLConn, config *configs.Config) *Authentication {
	return &Authentication{
		db:     db,
		config: config,
	}
}

func (a *Authentication) ValidateAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")

		client := &http.Client{}
		req, err := http.NewRequest("GET", a.config.App.AuthUrl, nil)
		if err != nil {
			response.WithMessage(w, http.StatusInternalServerError, err.Error())
			return
		}

		req.Header.Add("Authorization", tokenStr)

		resp, err := client.Do(req)
		if err != nil {
			response.WithMessage(w, http.StatusInternalServerError, err.Error())
			return
		}

		decoder := json.NewDecoder(resp.Body)
		var responseBody ValidateAuthResponse
		err = decoder.Decode(&responseBody)
		if err != nil {
			response.WithError(w, failure.BadRequest(err))
		}

		ctx := context.WithValue(r.Context(), "claims", responseBody.Data)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authentication) ClientCredential(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get(HeaderAuthorization)
		token := oauth.New(a.db.Read, oauth.Config{})

		parseToken, err := token.ParseWithAccessToken(accessToken)
		if err != nil {
			response.WithMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !parseToken.VerifyExpireIn() {
			response.WithMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authentication) ClientCredentialWithQueryParameter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		token := params.Get("token")
		tokenType := params.Get("token_type")
		accessToken := tokenType + " " + token

		auth := oauth.New(a.db.Read, oauth.Config{})
		parseToken, err := auth.ParseWithAccessToken(accessToken)
		if err != nil {
			response.WithMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !parseToken.VerifyExpireIn() {
			response.WithMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authentication) Password(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get(HeaderAuthorization)
		token := oauth.New(a.db.Read, oauth.Config{})

		parseToken, err := token.ParseWithAccessToken(accessToken)
		if err != nil {
			response.WithMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !parseToken.VerifyExpireIn() {
			response.WithMessage(w, http.StatusUnauthorized, err.Error())
			return
		}

		if !parseToken.VerifyUserLoggedIn() {
			response.WithMessage(w, http.StatusUnauthorized, oauth.ErrorInvalidPassword)
			return
		}

		next.ServeHTTP(w, r)
	})
}
