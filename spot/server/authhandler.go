package server

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kristofferostlund/spot/spot/auth"
	"github.com/kristofferostlund/spot/spot/config"
)

func (s *server) handleAuthentication() http.HandlerFunc {
	authenticator, state := auth.RedirectAuthenticator(
		config.ClientID,
		config.ClientSecret,
		config.RedirectURL,
	)

	auth.OpenAuthURL(authenticator, state)

	return func(w http.ResponseWriter, r *http.Request) {
		token, err := authenticator.Token(state, r)
		if err != nil {
			http.Error(w, "Failed to get token", http.StatusNotFound)

			return
		}

		client := auth.RedirectClient(authenticator, token)

		go func() {
			defer closeServer(*s)

			s.callback(client)
		}()

		if _, err = w.Write([]byte("OK")); err != nil {
			logrus.Fatal(err)
		}
	}
}
