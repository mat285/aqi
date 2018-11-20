package util

const (
	slackOpenQuote  = '“'
	slackCloseQuote = '”'
)

// SplitOnSpacePreserveQuotes splits the string on spaces preserving whitespace in quotes
func SplitOnSpacePreserveQuotes(str string) []string {
	ret := []string{}
	state := 0
	runes := []rune(str)
	curr := []rune{}
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if state == 0 {
			if r == ' ' {
				continue
			} else if isQuote(r) {
				state = 2
				continue
			}
			curr = append(curr, r)
			state = 1
		} else if state == 1 {
			if r == ' ' {
				ret = append(ret, string(curr))
				curr = []rune{}
				state = 0
				continue
			} else if isQuote(r) {
				state = 2
				continue
			} else {
				curr = append(curr, r)
			}
		} else if state == 2 {
			if r == ' ' {
				if len(curr) == 0 || curr[len(curr)-1] == r {
					continue
				}
				curr = append(curr, r)
			} else if isQuote(r) {
				state = 1
				continue
			} else {
				curr = append(curr, r)
			}
		}
	}
	if len(curr) > 0 {
		ret = append(ret, string(curr))
	}
	return ret
}

func isQuote(r rune) bool {
	return r == '"' || r == slackCloseQuote || r == slackOpenQuote
}
