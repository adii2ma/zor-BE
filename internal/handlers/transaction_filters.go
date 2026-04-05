package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"be-zor/internal/models"
)

func parseTransactionFilters(c *fiber.Ctx) (models.TransactionFilters, error) {
	dateFromRaw := strings.TrimSpace(c.Query("dateFrom"))
	dateToRaw := strings.TrimSpace(c.Query("dateTo"))
	category := strings.TrimSpace(c.Query("category"))
	typeRaw := models.TransactionType(strings.ToLower(strings.TrimSpace(c.Query("type"))))

	var filters models.TransactionFilters

	if dateFromRaw != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromRaw)
		if err != nil {
			return models.TransactionFilters{}, errors.New("dateFrom must be in YYYY-MM-DD format")
		}
		filters.DateFrom = &dateFrom
	}

	if dateToRaw != "" {
		dateTo, err := time.Parse("2006-01-02", dateToRaw)
		if err != nil {
			return models.TransactionFilters{}, errors.New("dateTo must be in YYYY-MM-DD format")
		}
		filters.DateTo = &dateTo
	}

	if filters.DateFrom != nil && filters.DateTo != nil && filters.DateFrom.After(*filters.DateTo) {
		return models.TransactionFilters{}, errors.New("dateFrom cannot be later than dateTo")
	}

	if typeRaw != "" && typeRaw != models.TransactionTypeIncome && typeRaw != models.TransactionTypeExpense {
		return models.TransactionFilters{}, errors.New("type must be income or expense")
	}

	filters.Category = category
	filters.Type = typeRaw

	return filters, nil
}
