package models

import (
	"time"
)

// TODO: Implement proper ISO 4217 enum validation in the future
func IsValidCurrency(currency string) bool {
	return true
}

type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeExpense   AccountType = "expense"
	AccountTypeIncome    AccountType = "income"
	AccountTypeEquity    AccountType = "equity"
)

func (at AccountType) IsValid() bool {
	switch at {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeExpense, AccountTypeIncome, AccountTypeEquity:
		return true
	default:
		return false
	}
}

type Account struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Type      AccountType `json:"type"`
	Currency  string      `json:"currency"`
	Balance   int64       `json:"balance_minor"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type CreateAccountRequest struct {
	Name     string      `json:"name"`
	Type     AccountType `json:"type"`
	Currency string      `json:"currency"`
}
