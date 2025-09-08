package mail

import (
	"github.com/casbin/govaluate"
)

type CtxRecipients []*ctxRecipient

// Filter filters the recipients based on the provided filter expression.
// It evaluates the filter expression against each recipient's parameters
// and returns a slice of recipients that match the criteria.
//
// Parameters:
//
//	filter: A string representing the filter expression to evaluate.
//	         The expression should be in a format compatible with
//	         the govaluate library.
//
// Returns:
//
//	A slice of pointers to ctxRecipient that match the filter criteria,
//	or an error if the evaluation of the expression fails or if any
//	other error occurs during the filtering process.
func (cr CtxRecipients) Filter(filter string) ([]*ctxRecipient, error) {
	expression, err := govaluate.NewEvaluableExpression(filter)
	if err != nil {
		return nil, err
	}

	var filteredRecipients []*ctxRecipient
	for _, r := range cr {
		result, err := expression.Evaluate(r.Params)
		if err != nil {
			return nil, err
		}
		if result == true {
			filteredRecipients = append(filteredRecipients, r)
		}
	}
	return filteredRecipients, nil
}
