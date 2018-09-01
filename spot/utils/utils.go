package utils

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/zmb3/spotify"
)

const Numbers = "1234567890"

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

func OpenBrowser(url string) {
	var args []string

	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}

	cmd := exec.Command(args[0], append(args[1:], url)...)
	if err := cmd.Start(); err != nil {
		logrus.Fatalf("Failed to open browser: %v", err)
	}
}

func MakeStringSortable(input string, minNumberCount int) string {
	output := ""
	currentNumberRange := ""

	for _, character := range input {
		stringified := string(character)

		if strings.Contains(Numbers, stringified) {
			currentNumberRange += stringified

			continue
		}

		if currentNumberRange != "" {
			output += LeftPad(currentNumberRange, minNumberCount, "0")
			currentNumberRange = ""
		}

		output += stringified
	}

	if currentNumberRange != "" {
		return output + LeftPad(currentNumberRange, minNumberCount, "0")
	}

	return output
}

func LeftPad(input string, minWidth int, padChar string) string {
	if len(input) >= minWidth {
		return input
	}

	return MultiplyString(padChar, minWidth-len(input)) + input
}
