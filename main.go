package main

import (
	"fmt"
	"regexp/syntax"
	"strings"
)

type DescriptionRegex struct {
	Regex *syntax.Regexp
}

func main() {
	r, _ := MakeRegex("DTB_[0-9]{3}_[A-Z]{3}")

	fmt.Println(r.GetDescription())
}

func MakeRegex(regex string) (*DescriptionRegex, error) {
	r, err := syntax.Parse(regex, syntax.OneLine)
	if err != nil {
		return nil, err
	}

	return &DescriptionRegex{Regex: r}, nil
}

func (r *DescriptionRegex) GetDescription() string {
	return analyzeRegex(r.Regex)
}

func analyzeRegex(reg *syntax.Regexp) string {
	switch reg.Op {
	case syntax.OpEmptyMatch: // matches empty string
	case syntax.OpCharClass: // matches Runes interpreted as range pair list
		return getRuneDescription(reg.Rune, 2, " ODER ")
	case syntax.OpLiteral: // matches Runes sequence
		return getRuneDescription(reg.Rune, 1, "")
	case syntax.OpAnyCharNotNL: // matches any character except newline
		return "Beliebiges Zeichen ohne \\n"
	case syntax.OpAnyChar: // matches any character
		return "Beliebiges Zeichen"
	case syntax.OpBeginLine: // matches empty string at beginning of line
		return "Zeilenbeginn"
	case syntax.OpEndLine: // matches empty string at end of line
		return "Zeilenende"
	case syntax.OpBeginText: // matches empty string at beginning of text
		return "Textanfang"
	case syntax.OpEndText: // matches empty string at end of text
		return "Textende"
	case syntax.OpWordBoundary: // matches word boundary `\b`
		return "Wortanfang oder -ende"
	case syntax.OpNoWordBoundary: // matches word non-boundary `\B`
		return "Kein Wortanfang oder -ende"
	case syntax.OpCapture: // capturing subexpression
		capture := make([]string, len(reg.Sub))
		for idx, sub := range reg.Sub {
			capture[idx] = analyzeRegex(sub)
		}
		return strings.Join(capture, " ODER ")
	case syntax.OpStar: // matches Sub[0] any amount of times
		return getRepeatDescription(0, -1, analyzeRegex(reg.Sub[0]))
	case syntax.OpPlus: // matches at least one Sub[0]
		return getRepeatDescription(1, -1, analyzeRegex(reg.Sub[0]))
	case syntax.OpQuest: // matches 0 or 1 time Sub[0]
		return getRepeatDescription(0, 1, analyzeRegex(reg.Sub[0]))
	case syntax.OpRepeat: // matches Sub[0] between Min and Max times. When Max == -1 then Max should be infinite.
		return getRepeatDescription(reg.Min, reg.Max, analyzeRegex(reg.Sub[0]))
	case syntax.OpConcat: // matches concatenation of Subs
		concat := make([]string, len(reg.Sub))
		for idx, sub := range reg.Sub {
			concat[idx] = analyzeRegex(sub)
		}
		return strings.Join(concat, " + ")
	case syntax.OpAlternate: // matches alternation of Subs
		alternate := make([]string, len(reg.Sub))

		for idx, sub := range reg.Sub {
			alternate[idx] = analyzeRegex(sub)
		}
		return strings.Join(alternate, " ODER ")
	}
	return ""
}

func getRepeatDescription(min int, max int, char string) string {
	switch {
	case min == max:
		return fmt.Sprintf("%d-mal %s", min, char)
	case max == -1:
		return fmt.Sprintf("mindestens %d-mal %s", min, char)
	default:
		return fmt.Sprintf("zwischen %d und %d-mal %s", min, max, char)
	}
}

func getRuneDescription(runes []rune, groupSize int, join string) string {
	runeResult := make([]string, len(runes)/groupSize)
	for i, curRune := range runes {
		if i%groupSize != 0 {
			runeResult[i/groupSize] += " BIS "
		}
		runeResult[i/groupSize] += string(curRune)
	}
	return strings.Join(runeResult, join)
}

func findArguments(instructions []syntax.Inst, origin int, end int) []string {
	result := make([]string, 0)
	idx := int(instructions[origin].Out)
	current := ""
	for idx != origin {
		instr := instructions[idx]

		current += instr.String()
	}
	return result
}
