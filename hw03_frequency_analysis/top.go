// Package hw03frequencyanalysis implements a function that analyzes the frequency
// of words in a given text and returns the top 10 most frequent words.
package hw03frequencyanalysis

import (
	"regexp"
	"slices"
	"strings"
)

var (
	// Stripping the word of punctuation chars only when it has a single one at the start or end of the string.
	//nolint:lll
	strippedWordPattern = regexp.MustCompile(`^[\p{P}]?(?P<pat1>[^\p{P}]+.*[^\p{P}]+|[^\p{P}]|[^\p{P}]+.*)[\p{P}]?$|^(?P<pat2>[\p{P}]{2,}.*[^\p{P}]+)[\p{P}]?$`)
	singlePunctPattern  = regexp.MustCompile(`^[\p{P}]$`)
)

func parseWord(word string) string {
	if isMatched := singlePunctPattern.Match([]byte(word)); isMatched {
		return ""
	}

	// Stripping and lowercasing the word if it matches the punctuation pattern.
	matches := strippedWordPattern.FindStringSubmatch(word)
	if len(matches) != 0 {
		for i, matchName := range strippedWordPattern.SubexpNames() {
			if (matchName == "pat1" || matchName == "pat2") && matches[i] != "" {
				return strings.ToLower(matches[i])
			}
		}
	}

	return strings.ToLower(word)
}

// Top10 analyzes the frequency of words in the input string `sourceText`
// and returns a slice containing the top 10 most frequent words. If multiple
// words have the same frequency, they are sorted lexicographically. If the
// input string is empty, it returns nil.
func Top10(sourceText string) []string {
	if sourceText == "" {
		return nil
	}

	const topCount = 10
	wordFreqs := make(map[string]int)

	textByWords := strings.Fields(sourceText)
	words := make([]string, 0, len(textByWords))

	// Parsing goes here.
	for _, word := range textByWords {
		parsedWord := parseWord(word)
		if parsedWord == "" {
			continue
		}
		// Filling words slice with unique words.
		if _, ok := wordFreqs[parsedWord]; !ok {
			words = append(words, parsedWord)
		}
		wordFreqs[parsedWord]++
	}

	// No words found in the text.
	if len(wordFreqs) == 0 {
		return nil
	}

	// Sorting the slice both by frequency and lexicographically.
	slices.SortFunc(words, func(a, b string) int {
		if wordFreqs[a] == wordFreqs[b] {
			return strings.Compare(a, b)
		}
		return wordFreqs[b] - wordFreqs[a]
	})

	res := make([]string, min(topCount, len(words)))
	_ = copy(res, words)

	return res
}
