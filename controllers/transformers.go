package controllers

import (
	"time"

	"github.com/trtstm/budgetr/models"
)

// CategoryResponse holds the response data for a category.
type CategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// TransformCategory transforms one or more categories.
func TransformCategory(categories ...*models.Category) (result []*CategoryResponse) {
	result = []*CategoryResponse{}
	for _, category := range categories {
		resp := &CategoryResponse{
			ID:   category.ID,
			Name: category.Name,
		}
		result = append(result, resp)
	}

	return
}

// ExpenditureResponse holds the response data for an expenditure.
type ExpenditureResponse struct {
	ID       uint              `json:"id"`
	Amount   float64           `json:"amount"`
	Date     time.Time         `json:"date"`
	Category *CategoryResponse `json:"category"`
}

// TransformExpenditure transforms one or more expenditures.
func TransformExpenditure(expenditures ...*models.Expenditure) (result []*ExpenditureResponse) {
	result = []*ExpenditureResponse{}
	for _, expenditure := range expenditures {
		resp := &ExpenditureResponse{
			ID:     expenditure.ID,
			Amount: expenditure.Amount,
			Date:   expenditure.Date,
		}

		if expenditure.Category != nil {
			resp.Category = TransformCategory(expenditure.Category)[0]
		}

		result = append(result, resp)
	}

	return
}
