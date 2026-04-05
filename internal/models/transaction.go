package models

import (
	"time"

	"github.com/uptrace/bun"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type Transaction struct {
	ID              string          `json:"id"`
	UserID          string          `json:"userId,omitempty"`
	Amount          float64         `json:"amount"`
	Type            TransactionType `json:"type"`
	Category        string          `json:"category"`
	TransactionDate time.Time       `json:"transactionDate"`
	Description     string          `json:"description,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type TransactionRecord struct {
	bun.BaseModel `bun:"table:transactions,alias:t"`

	ID              string          `bun:"id,pk" json:"-"`
	UserID          string          `bun:"user_id,notnull" json:"-"`
	Amount          float64         `bun:"amount,notnull" json:"-"`
	Type            TransactionType `bun:"type,notnull" json:"-"`
	Category        string          `bun:"category,notnull" json:"-"`
	TransactionDate time.Time       `bun:"transaction_date,notnull" json:"-"`
	Description     string          `bun:"description" json:"-"`
	CreatedAt       time.Time       `bun:"created_at,notnull" json:"-"`
	UpdatedAt       time.Time       `bun:"updated_at,notnull" json:"-"`
}

type CategorySummary struct {
	Category         string  `json:"category"`
	Income           float64 `json:"income"`
	Expense          float64 `json:"expense"`
	NetBalance       float64 `json:"netBalance"`
	TransactionCount int     `json:"transactionCount"`
}

type TrendSummary struct {
	Period     string    `json:"period"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	Income     float64   `json:"income"`
	Expense    float64   `json:"expense"`
	NetBalance float64   `json:"netBalance"`
}

type DashboardSummary struct {
	TotalIncome    float64           `json:"totalIncome"`
	TotalExpenses  float64           `json:"totalExpenses"`
	NetBalance     float64           `json:"netBalance"`
	CategoryTotals []CategorySummary `json:"categoryTotals"`
	RecentActivity []Transaction     `json:"recentActivity"`
	MonthlyTrends  []TrendSummary    `json:"monthlyTrends"`
	WeeklyTrends   []TrendSummary    `json:"weeklyTrends"`
}

func (t TransactionRecord) ToTransaction() Transaction {
	return Transaction{
		ID:              t.ID,
		UserID:          t.UserID,
		Amount:          t.Amount,
		Type:            t.Type,
		Category:        t.Category,
		TransactionDate: t.TransactionDate,
		Description:     t.Description,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
