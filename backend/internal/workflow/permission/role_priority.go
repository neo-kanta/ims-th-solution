package permission

// RolePriority orders group codes from least to most senior. Used to
// snapshot a single canonical role onto an approval record for audit.
// Codes match the permissions module's group codes.
var RolePriority = []string{
	"TRADER",
	"FUND_MANAGER",
	"RISK_CONTROL",
	"ADMIN",
}

// PickHighest returns the most senior role present in roles, or "" when
// none of them appears in RolePriority. O(n*m) is fine — both lists
// are short.
func PickHighest(roles []string) string {
	highest := ""
	highestRank := -1
	for _, r := range roles {
		for rank, known := range RolePriority {
			if r == known && rank > highestRank {
				highestRank = rank
				highest = r
			}
		}
	}
	return highest
}
