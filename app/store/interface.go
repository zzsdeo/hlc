package store

import "hlc/app/models"

type AccountsRepository interface {
	FilterAccounts(FilterQuery) (FilterResult, error)
	GroupAccounts(GroupsQuery) (GroupsResult, error)
	CreateAccounts([]models.Account) (int, error)
}

type FilterQuery struct {
	SexEq             string   `json:"sex_eq"` //m/f
	EmailDomain       string   `json:"email_domain"`
	EmailLt           string   `json:"email_lt"`
	EmailGt           string   `json:"email_gt"`
	StatusEq          string   `json:"status_eq"`
	StatusNeq         string   `json:"status_neq"`
	FNameEq           string   `json:"fname_eq"`
	FNameAny          string   `json:"fname_any"`
	FNameNull         string   `json:"fname_null"`
	SNameEq           string   `json:"sname_eq"`
	SNameStarts       string   `json:"sname_starts"`
	SNameNull         string   `json:"sname_null"`
	PhoneCode         string   `json:"phone_code"`
	PhoneNull         string   `json:"phone_null"`
	CountryEq         string   `json:"country_eq"`
	CountryNull       string   `json:"country_null"`
	CityEq            string   `json:"city_eq"`
	CityAny           string   `json:"city_any"`
	CityNull          string   `json:"city_null"`
	BirthLt           int      `json:"birth_lt"`
	BirthGt           int      `json:"birth_gt"`
	BirthYear         int      `json:"birth_year"`
	InterestsContains []string `json:"interests_contains"`
	InterestsAny      []string `json:"interests_any"`
	LikesContains     []uint32 `json:"likes_contains"`
	PremiumNow        int      `json:"premium_now"`
	PremiumNull       int      `json:"premium_null"`
	Limit             uint     `json:"limit"`
}

type FilterResult struct {
	Accounts []models.Account `json:"accounts"`
}

type GroupsQuery struct {
	models.Account
	Order int    `json:"order"`
	Limit uint   `json:"limit"`
	Keys  string `json:"keys"`
}

type GroupsResult struct {
	Groups []Group `json:"groups"`
}

type Group struct {
	Keys
	Count int `json:"count"`
}

type Keys map[string]string
