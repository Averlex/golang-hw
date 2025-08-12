// Package hw10programoptimization the function which needs to be optimized.
package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// User contains user data.
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

// DomainStat contains the number of users in each domain.
type DomainStat map[string]int

// GetDomainStat returns the number of users in each domain.
func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	res, err := processByLine(r, domain)
	if err != nil {
		return nil, fmt.Errorf("processing error: %w", err)
	}
	return res, nil
}

func processByLine(r io.Reader, domain string) (map[string]int, error) {
	scanner := bufio.NewScanner(r)
	result := make(map[string]int)

	expression, err := regexp.Compile("\\." + domain)
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		line := scanner.Bytes()
		var user User
		if err := user.UnmarshalJSON(line); err != nil {
			return nil, err
		}
		parsedDomain, count, err := countDomains(&user, expression)
		if err != nil {
			return nil, err
		}
		if parsedDomain == "" {
			continue
		}
		result[parsedDomain] += count
	}

	return result, nil
}

func countDomains(user *User, expression *regexp.Regexp) (string, int, error) {
	count := 0
	var parsedDomain string
	if user == nil || expression == nil {
		return parsedDomain,
			count,
			fmt.Errorf("not enough data passed for processing: user=%v, expression=%v", user, expression)
	}

	matched := expression.Match([]byte(user.Email))

	if matched {
		parsedDomain = strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
		count++
	}
	return parsedDomain, count, nil
}
