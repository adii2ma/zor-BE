package summary

import (
	"sort"

	"be-zor/internal/models"
)

func BuildOrganizationSummary(transactions []models.AdminTransaction) models.DashboardSummary {
	normalized := make([]models.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		normalized = append(normalized, transaction.ToTransaction())
	}

	return BuildDashboardSummary(normalized)
}

func BuildAnalystUserTransactions(transactions []models.AdminTransaction) []models.AnalystUserTransactions {
	type bucket struct {
		userName      string
		userEmail     string
		totalIncome   float64
		totalExpenses float64
		transactions  []models.AnalystTransaction
	}

	grouped := make(map[string]*bucket)
	for _, transaction := range transactions {
		key := transaction.UserEmail
		entry, ok := grouped[key]
		if !ok {
			entry = &bucket{
				userName:  transaction.UserName,
				userEmail: transaction.UserEmail,
			}
			grouped[key] = entry
		}

		if transaction.Type == models.TransactionTypeIncome {
			entry.totalIncome += transaction.Amount
		} else if transaction.Type == models.TransactionTypeExpense {
			entry.totalExpenses += transaction.Amount
		}

		entry.transactions = append(entry.transactions, transaction.ToAnalystTransaction())
	}

	recordsByUser := make([]models.AnalystUserTransactions, 0, len(grouped))
	for _, entry := range grouped {
		sort.Slice(entry.transactions, func(i, j int) bool {
			if entry.transactions[i].TransactionDate.Equal(entry.transactions[j].TransactionDate) {
				return entry.transactions[i].CreatedAt.After(entry.transactions[j].CreatedAt)
			}
			return entry.transactions[i].TransactionDate.After(entry.transactions[j].TransactionDate)
		})

		recordsByUser = append(recordsByUser, models.AnalystUserTransactions{
			UserName:         entry.userName,
			UserEmail:        entry.userEmail,
			TransactionCount: len(entry.transactions),
			TotalIncome:      entry.totalIncome,
			TotalExpenses:    entry.totalExpenses,
			NetBalance:       entry.totalIncome - entry.totalExpenses,
			Transactions:     entry.transactions,
		})
	}

	sort.Slice(recordsByUser, func(i, j int) bool {
		if recordsByUser[i].UserName == recordsByUser[j].UserName {
			return recordsByUser[i].UserEmail < recordsByUser[j].UserEmail
		}
		return recordsByUser[i].UserName < recordsByUser[j].UserName
	})

	return recordsByUser
}
