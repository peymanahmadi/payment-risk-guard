package risk

import (
	"context"
	"fmt"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
)

type Verdict struct {
	Score     float64
	Triggered bool
	Reasons   []string
	Results   []Result
}

type Engine struct {
	rules     []Rule
	threshold float64
}

func NewEngine(threshold float64, rules ...Rule) *Engine {
	return &Engine{rules: rules, threshold: threshold}
}

func (e *Engine) Evaluate(ctx context.Context, tx *domain.Transaction) (Verdict, error) {
	var (
		maxScore float64
		reasons  []string
		results  = make([]Result, 0, len(e.rules))
	)

	for _, rule := range e.rules {
		res, err := rule.Evaluate(ctx, tx)
		if err != nil {
			return Verdict{}, fmt.Errorf("rule %q: %w", rule.Name(), err)
		}
		results = append(results, res)

		if res.Triggered {
			reasons = append(reasons, fmt.Sprintf("[%s] %s", res.RuleName, res.Reason))
			if res.Score > maxScore {
				maxScore = res.Score
			}
		}
	}

	return Verdict{
		Score:     maxScore,
		Triggered: maxScore >= e.threshold,
		Reasons:   reasons,
		Results:   results,
	}, nil
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
