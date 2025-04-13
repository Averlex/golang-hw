// Package hw03frequencyanalysis implements a function that analyzes the frequency
// of words in a given text and returns the top 10 most frequent words.
package hw03frequencyanalysis

import (
	"slices"
	"strings"
)

// TODO: or []*string? To avoid reallocations.
type freqCounter struct {
	Count int
	Words []string
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
		// TODO: Some text parsing logic
		wordFreqs[word]++
	}

	// No words found in the text.
	if len(wordFreqs) == 0 {
		return nil
	}

	freqs := make(map[int]*freqCounter, len(wordFreqs))

	// Grouping up words by their frequency.
	for k, v := range wordFreqs {
		if _, ok := freqs[v]; !ok {
			freqs[v] = &freqCounter{v, make([]string, 0, len(textByWords))}
		}
		freqs[v].Words = append(freqs[v].Words, k)
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
		slices.Sort(v.Words) // Sorting lexicographically within the given frequency group.
		for _, word := range v.Words {
			if count == topCount {
				break
			}
			res = append(res, word)
			count++
		}
	}

	return res
}
