package lsportal

import (
	"regexp"
	"sort"
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Process the text of the forwarder to replace anything except newlines not within the regexs with a space
// inclusionRegex: A multiline regex that should match the text you want to keep within its first match group, it is expected to match many times
// exclusionRegex: A multiline regex that should match text you want remove from within an inclusion
// returns the new text and a slice of ranges for the inclusions
func whitespaceExceptInclusions(text string, inclusionRegex string, exclusionRegex string) (string, []protocol.Range) {
	// Compile the inclusion regex
	incRegex := regexp.MustCompile(inclusionRegex)

	// Convert the text to a rune slice
	runes := []rune(text)

	// Create a slice to store the result
	result := make([]rune, len(runes))
	var lineEnds []int
	// Initialize the result slice with spaces
	for i := range result {
		if runes[i] == '\n' {
			lineEnds = append(lineEnds, i)
			result[i] = '\n'
		} else {
			result[i] = ' '
		}
	}

	var ranges []protocol.Range
	// Find all matches of the inclusion regex
	matches := incRegex.FindAllStringSubmatchIndex(text, -1)
	// Iterate over the matches
	for _, match := range matches {
		// Check if there is a capturing group
		if len(match) >= 4 {
			start, end := match[2], match[3]
			ranges = append(ranges, getRange(start, end, lineEnds))

			// Copy the captured text to the result slice
			copy(result[start:end], runes[start:end])
		}
	}

	// Convert the result slice back to a string
	processedText := string(result)

	// Compile the exclusion regex if provided
	if exclusionRegex != "" {
		excRegex := regexp.MustCompile(exclusionRegex)
		// Replace any matches of the exclusion regex with spaces
		processedText = excRegex.ReplaceAllStringFunc(processedText, func(match string) string {
			return strings.Map(func(r rune) rune {
				if r == '\n' {
					return '\n'
				}
				return ' '
			}, match)
		})
	}

	return processedText, ranges
}

func getPosition(offset int, lineEnds []int) protocol.Position {
	line := sort.Search(len(lineEnds), func(i int) bool {
		return lineEnds[i] >= offset
	})
	character := offset
	if line > 0 {
		character -= lineEnds[line-1] + 1
	}
	return protocol.Position{
		Line:      protocol.UInteger(line),
		Character: protocol.UInteger(character),
	}
}

func getRange(start int, end int, lineEnds []int) protocol.Range {
	startPos := getPosition(start, lineEnds)
	endPos := getPosition(end, lineEnds)
	return protocol.Range{
		Start: startPos,
		End:   endPos,
	}
}
