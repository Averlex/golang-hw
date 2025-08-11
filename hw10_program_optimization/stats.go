// Package hw10programoptimization implements the task for a following homework.
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
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

func getUsers(r io.Reader) ([]User, error) {
	result := make([]User, 0)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		var user User
		if err := user.UnmarshalJSON(line); err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	return result, nil
}

func countDomains(u []User, domain string) (DomainStat, error) {
	result := make(DomainStat)
	expression, err := regexp.Compile("\\." + domain)
	if err != nil {
		return nil, err
	}

	for _, user := range u {
		matched := expression.Match([]byte(user.Email))

		if matched {
			domain := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
			num := result[domain]
			num++
			result[domain] = num
		}
	}
	return result, nil
}
