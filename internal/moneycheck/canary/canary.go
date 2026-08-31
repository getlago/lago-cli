//go:build moneycheck_canary

// Package canary holds deliberate monetary-float defects. It is excluded from every
// normal build by its tag and exists only so TestCanaryIsRejected can prove the
// checker still fails on the defect classes it is supposed to catch. A linter with
// no negative test is a linter nobody notices is broken.
package canary

type Invoice struct {
	AmountCents   *float64           // pointer: optional JSON field
	TotalDueCents []float64          // slice
	FeeByCurrency map[string]float64 // map value
	Cents         float64            // bare vocabulary term
	Rate          float64
	Charge        float64
	Discount      float64
	VatAmount     int64 // legitimate: must not be reported
}

type Money float64 // type declaration

var totalAmount float64 = 12.34 // var declaration

func TotalAmount() float64 { return 0.1 + 0.2 } // unnamed float result

func AttemptCount() float64 { return 1 } // not monetary: must not be reported
