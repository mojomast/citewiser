package hygiene

import "github.com/mojomast/citewiser/pkg/packer"

func signalForScore(score, threshold float64) packer.HygieneSignal {
	if score < threshold {
		return packer.HygieneRed
	}
	if score < 0.70 {
		return packer.HygieneYellow
	}
	return packer.HygieneGreen
}
