package summary

import (
	"fmt"
	"sort"
	"time"

	"be-zor/internal/models"
)

const defaultRecentActivityLimit = 5

func BuildDashboardSummary(transactions []models.Transaction) models.DashboardSummary {
	totalIncome := CalculateTotalIncome(transactions)
	totalExpenses := CalculateTotalExpenses(transactions)

	return models.DashboardSummary{
		TotalIncome:    totalIncome,
		TotalExpenses:  totalExpenses,
		NetBalance:     CalculateNetBalance(totalIncome, totalExpenses),
		CategoryTotals: CalculateCategoryTotals(transactions),
		RecentActivity: CalculateRecentActivity(transactions, defaultRecentActivityLimit),
		MonthlyTrends:  CalculateMonthlyTrends(transactions),
		WeeklyTrends:   CalculateWeeklyTrends(transactions),
	}
}

func CalculateTotalIncome(transactions []models.Transaction) float64 {
	var total float64
	for _, transaction := range transactions {
		if transaction.Type == models.TransactionTypeIncome {
			total += transaction.Amount
		}
	}

	return total
}

func CalculateTotalExpenses(transactions []models.Transaction) float64 {
	var total float64
	for _, transaction := range transactions {
		if transaction.Type == models.TransactionTypeExpense {
			total += transaction.Amount
		}
	}

	return total
}

func CalculateNetBalance(totalIncome, totalExpenses float64) float64 {
	return totalIncome - totalExpenses
}

func CalculateCategoryTotals(transactions []models.Transaction) []models.CategorySummary {
	type bucket struct {
		income  float64
		expense float64
		count   int
	}

	grouped := make(map[string]*bucket)
	for _, transaction := range transactions {
		entry, ok := grouped[transaction.Category]
		if !ok {
			entry = &bucket{}
			grouped[transaction.Category] = entry
		}

		entry.count++
		switch transaction.Type {
		case models.TransactionTypeIncome:
			entry.income += transaction.Amount
		case models.TransactionTypeExpense:
			entry.expense += transaction.Amount
		}
	}

	categories := make([]models.CategorySummary, 0, len(grouped))
	for category, totals := range grouped {
		categories = append(categories, models.CategorySummary{
			Category:         category,
			Income:           totals.income,
			Expense:          totals.expense,
			NetBalance:       totals.income - totals.expense,
			TransactionCount: totals.count,
		})
	}

	sort.Slice(categories, func(i, j int) bool {
		if categories[i].NetBalance == categories[j].NetBalance {
			return categories[i].Category < categories[j].Category
		}
		return categories[i].NetBalance > categories[j].NetBalance
	})

	return categories
}

func CalculateRecentActivity(transactions []models.Transaction, limit int) []models.Transaction {
	if limit <= 0 || len(transactions) == 0 {
		return []models.Transaction{}
	}

	activity := append([]models.Transaction(nil), transactions...)
	sort.Slice(activity, func(i, j int) bool {
		if activity[i].TransactionDate.Equal(activity[j].TransactionDate) {
			return activity[i].CreatedAt.After(activity[j].CreatedAt)
		}
		return activity[i].TransactionDate.After(activity[j].TransactionDate)
	})

	if len(activity) > limit {
		activity = activity[:limit]
	}

	return activity
}

func CalculateMonthlyTrends(transactions []models.Transaction) []models.TrendSummary {
	return calculateTrends(transactions, startOfMonth, endOfMonth, func(date time.Time) string {
		return date.UTC().Format("2006-01")
	})
}

func CalculateWeeklyTrends(transactions []models.Transaction) []models.TrendSummary {
	return calculateTrends(transactions, startOfWeek, endOfWeek, func(date time.Time) string {
		year, week := date.UTC().ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	})
}

func calculateTrends(
	transactions []models.Transaction,
	startFn func(time.Time) time.Time,
	endFn func(time.Time) time.Time,
	labelFn func(time.Time) string,
) []models.TrendSummary {
	type bucket struct {
		startDate time.Time
		endDate   time.Time
		income    float64
		expense   float64
	}

	grouped := make(map[string]*bucket)
	for _, transaction := range transactions {
		startDate := startFn(transaction.TransactionDate)
		key := labelFn(startDate)

		entry, ok := grouped[key]
		if !ok {
			entry = &bucket{
				startDate: startDate,
				endDate:   endFn(startDate),
			}
			grouped[key] = entry
		}

		switch transaction.Type {
		case models.TransactionTypeIncome:
			entry.income += transaction.Amount
		case models.TransactionTypeExpense:
			entry.expense += transaction.Amount
		}
	}

	trends := make([]models.TrendSummary, 0, len(grouped))
	for period, totals := range grouped {
		trends = append(trends, models.TrendSummary{
			Period:     period,
			StartDate:  totals.startDate,
			EndDate:    totals.endDate,
			Income:     totals.income,
			Expense:    totals.expense,
			NetBalance: totals.income - totals.expense,
		})
	}

	sort.Slice(trends, func(i, j int) bool {
		return trends[i].StartDate.Before(trends[j].StartDate)
	})

	return trends
}

func startOfMonth(date time.Time) time.Time {
	date = date.UTC()
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func endOfMonth(date time.Time) time.Time {
	start := startOfMonth(date)
	return start.AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func startOfWeek(date time.Time) time.Time {
	date = date.UTC()
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).
		AddDate(0, 0, -(weekday - 1))
}

func endOfWeek(date time.Time) time.Time {
	start := startOfWeek(date)
	return start.AddDate(0, 0, 7).Add(-time.Nanosecond)
}
