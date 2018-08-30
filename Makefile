
include .env
export SPOTIFY_ID
export SPOTIFY_SECRET

local:
	go run spot/cli/cli.go -user drklump
