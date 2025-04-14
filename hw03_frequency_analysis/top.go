// Package hw03frequencyanalysis implements a function that analyzes the frequency
// of words in a given text and returns the top 10 most frequent words.
package hw03frequencyanalysis

import (
	"regexp"
	"slices"
	"strings"
)

type freqCounter struct {
	Count int
	Words []*string // To avoid unnecessary reallocations.
}

var (
	strippedWordPattern = regexp.MustCompile(`^[\p{P}]?(?P<strippedWord>[^\p{P}]+.*[^\p{P}]+)[\p{P}]?$`)
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
			if matchName == "strippedWord" {
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

	// The worst case scenario -> sourceText = "a a a a ...".
	textByWords := strings.Fields(sourceText)

	for _, word := range textByWords {
		parsedWord := parseWord(word)
		if parsedWord == "" {
			continue
		}
		wordFreqs[parsedWord]++
	}

	// No words found in the text.
	if len(wordFreqs) == 0 {
		return nil
	}

	freqs := make(map[int]*freqCounter, len(wordFreqs))

	// Grouping up words by their frequency.
	for k, v := range wordFreqs {
		if _, ok := freqs[v]; !ok {
			freqs[v] = &freqCounter{v, make([]*string, 0, len(textByWords))}
		}
		clonedString := strings.Clone(k)
		freqs[v].Words = append(freqs[v].Words, &clonedString)
	}

	// Flattening frequencies to a simple slice.
	flattenedFreqs := make([]*freqCounter, 0, len(freqs))
	for _, v := range freqs {
		flattenedFreqs = append(flattenedFreqs, v)
	}

	// Sorting groups of words by each group's frequency.
	slices.SortFunc(flattenedFreqs, func(a, b *freqCounter) int { return b.Count - a.Count })

	count := 0
	res := make([]string, 0, topCount)
	for _, v := range flattenedFreqs {
		// Sorting lexicographically within the given frequency group.
		slices.SortFunc(v.Words, func(a, b *string) int { return strings.Compare(*a, *b) })
		for _, word := range v.Words {
			if count == topCount {
				break
			}
			res = append(res, *word)
			count++
		}
	}

	return res
}
