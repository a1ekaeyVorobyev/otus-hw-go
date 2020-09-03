package hw10_program_optimization //nolint:golint,stylecheck

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type User struct {
	ID       int    `json:"-"`
	Name     string `json:"-"`
	Username string `json:"-"`
	Email    string
	Phone    string `json:"-"`
	Password string `json:"-"`
	Address  string `json:"-"`
}

type DomainStat map[string]int

var (
	userCount = 0
	user      User
	result    = make(DomainStat)
)

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %s", err)
	}

	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (users, error) {
	var result users
	scanner := bufio.NewScanner(r)
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	userCount = 0
	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &user)
		if err != nil {
			return result, err
		}
		result[userCount] = user
		userCount++
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}

	return result, nil
}

func countDomains(u users, domain string) (DomainStat, error) {
	result = make(DomainStat)
	if userCount == 0 {
		userCount = 100_000
	}
	cnt := 0
	for _, user := range u[:userCount] {
		cnt = len(user.Email) - len(domain)
		if cnt > 0 && user.Email[cnt:] == domain {
			if i := strings.LastIndex(user.Email, "@"); i > 0 {
				result[strings.ToLower(user.Email[i+1:])]++
			}
		}
	}

	return result, nil
}
