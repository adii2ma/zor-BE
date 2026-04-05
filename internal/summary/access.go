package summary

import (
	"sort"
	"strconv"

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
		totalIncome   float64
		totalExpenses float64
		transactions  []models.AnalystTransaction
	}

	grouped := make(map[string]*bucket)
	for _, transaction := range transactions {
		key := transaction.UserID
		entry, ok := grouped[key]
		if !ok {
			entry = &bucket{}
			grouped[key] = entry
		}

		if transaction.Type == models.TransactionTypeIncome {
			entry.totalIncome += transaction.Amount
		} else if transaction.Type == models.TransactionTypeExpense {
			entry.totalExpenses += transaction.Amount
		}

		entry.transactions = append(entry.transactions, transaction.ToAnalystTransaction())
	}

	userIDs := make([]string, 0, len(grouped))
	for userID := range grouped {
		userIDs = append(userIDs, userID)
	}
	sort.Strings(userIDs)

	recordsByUser := make([]models.AnalystUserTransactions, 0, len(grouped))
	for index, userID := range userIDs {
		entry := grouped[userID]
		sort.Slice(entry.transactions, func(i, j int) bool {
			if entry.transactions[i].TransactionDate.Equal(entry.transactions[j].TransactionDate) {
				return entry.transactions[i].CreatedAt.After(entry.transactions[j].CreatedAt)
			}
			return entry.transactions[i].TransactionDate.After(entry.transactions[j].TransactionDate)
		})

		recordsByUser = append(recordsByUser, models.AnalystUserTransactions{
			AccountLabel:     accountLabel(index + 1),
			TransactionCount: len(entry.transactions),
			TotalIncome:      entry.totalIncome,
			TotalExpenses:    entry.totalExpenses,
			NetBalance:       entry.totalIncome - entry.totalExpenses,
			Transactions:     entry.transactions,
		})
	}

	return recordsByUser
}

func accountLabel(index int) string {
	return "Account " + strconv.Itoa(index)
}
