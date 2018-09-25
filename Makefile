
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

cli-redirect-recommendation:
	go run cli/cli.go \
		-user drklump \
		-playlist-pattern '^Metal [0-9]+' \
		-credentials-flow redirect \
		-output-type console \
		-operation recommendation

cli-redirect-check-track:
	go run cli/cli.go \
		-user drklump \
		-playlist-pattern '^Metal [0-9]+' \
		-credentials-flow redirect \
		-output-type console \
		-operation check-track

cli-redirect-check-playlist-holes:
	go run cli/cli.go \
		-user drklump \
		-playlist-pattern '^Metal [0-9]+' \
		-credentials-flow redirect \
		-output-type console \
		-operation check-playlist-holes

.PHONY: cli server
