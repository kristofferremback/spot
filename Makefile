
include .env
export SPOTIFY_ID
export SPOTIFY_SECRET

cli:
	go run cli/cli.go -user drklump

cli-redirect:
	go run cli/cli.go -user drklump -credentials-flow redirect -output-type playlist

.PHONY: cli server
