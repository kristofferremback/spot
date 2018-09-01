
include .env
export SPOTIFY_ID
export SPOTIFY_SECRET

cli:
	go run cli/cli.go \
		-user drklump \
		-playlist-pattern '^Metal [0-9]+'

cli-redirect:
	go run cli/cli.go \
		-user drklump \
		-playlist-pattern '^Metal [0-9]+' \
		-credentials-flow redirect \
		-output-type playlist

.PHONY: cli server
