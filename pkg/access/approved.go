package access

func HasTrustedApprover(ctx Context, approvedBy []string) bool {
	trusted := map[string]bool{}
	for _, principal := range ctx.TrustedApprovers {
		trusted[principal] = true
	}
	for _, principal := range approvedBy {
		if trusted[principal] {
			return true
		}
	}
	return false
}
