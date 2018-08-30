package utils

import (
	"strings"

	"github.com/zmb3/spotify"
)

func JoinArtists(artists []spotify.SimpleArtist, separator string) string {
	return strings.Join(
		func() []string {
			output := []string{}
			for _, a := range artists {
				output = append(output, a.Name)
			}
			return output
		}(),
		separator,
	)
}

func ChunkIDs(ids []spotify.ID, chunkSize int) [][]spotify.ID {
	chunks := [][]spotify.ID{[]spotify.ID{}}

	for _, id := range ids {
		chunkIndex := len(chunks) - 1

		if len(chunks[chunkIndex]) < chunkSize {
			chunks[chunkIndex] = append(chunks[chunkIndex], id)
		} else {
			chunks = append(chunks, []spotify.ID{id})
		}
	}

	return chunks
}

func FixedWidthString(input string, length int) string {
	dots := "..."

	if len(input) > length {
		input = input[:length-len(dots)] + dots
	}

	output := ""
	for i := 0; i < length; i++ {
		if len(input) > i {
			output += string(input[i])
		} else {
			output += " "
		}
	}

	return output
}

func MultiplyString(value string, iterations int) string {
	output := ""

	for i := 0; i < iterations; i++ {
		output += value
	}

	return output
}
