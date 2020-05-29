package main

import (
	"fmt"
	"reflect"
	"regexp/syntax"
	"strconv"
	"strings"
)

// DescriptionRegex represents a Regex Struct for which a description can be generated
type DescriptionRegex struct {
	Regex  *syntax.Regexp
	Config DescriptionRegexConfig
}

// DescriptionRegexConfig represents a configuration struct for Translation and Syntax configurations when generating descriptions
type DescriptionRegexConfig struct {
	Translation TranslationConfig
	Syntax      SyntaxConfig
}

// TranslationConfig represents a configuration struct for Translation of certain keywords when generating descriptions
type TranslationConfig struct {
	And            string // Translation for And.
	Or             string // Translation for Or.
	Between        string // Translation for Between. E.g. <BETWEEN> x and y.
	Without        string // Translation for Without. E.g. Any Char <WITHOUT> \n.
	AnyChar        string // Translation for any possible Char.
	BeginLine      string // Translation for the beginning of a line .
	EndLine        string // Translation for the end of a line.
	BeginText      string // Translation for the beginning of a text.
	EndText        string // Translation for the end of a text.
	WordBoundary   string // Translation for the start of a word.
	NoWordBoundary string // Translation for the end of a word.
	AtLeast        string // Translation for at least. E.g. <AT LEAST> x times.
	Times          string // Translation for times. E.g. x-<TIMES> y.
	To             string // Translation for to. E.g. Any number from 5 <TO> 9.
	From           string // Translation for from. E.g. Any number <FROM> 5 to 9.
	Whitespace     string `special:"32"` // Translation for whitespace.
}

// SyntaxConfig represents a configuration struct for Syntax configurations
type SyntaxConfig struct {
	UseComma bool   // Determines whether commas should be used to shorten the Description. E.g. if true (A|B|C) will be A, B <or> C instead of A <OR> B <OR> C.
	Indent   string // Determines the indent between individual statements of the Regex Description.
}

// GermanTranslationConfig represents the default TranslationConfig in german.
var GermanTranslationConfig = TranslationConfig{
	And:            "und",
	Or:             "oder",
	Between:        "zwischen",
	AnyChar:        "Beliebiges Zeichen",
	BeginLine:      "Zeilenbeginn",
	EndLine:        "Zeilenende",
	BeginText:      "Textanfang",
	EndText:        "Textende",
	WordBoundary:   "Wortgrenze",
	NoWordBoundary: "Keine Wortgrenze",
	AtLeast:        "Mindestens",
	Times:          "mal",
	To:             "Bis",
	From:           "von",
	Whitespace:     "Leerzeichen",
}

// EnglishTranslationConfig represents the default TranslationConfig in english.
var EnglishTranslationConfig = TranslationConfig{
	And:            "and",
	Or:             "or",
	Between:        "between",
	AnyChar:        "any char",
	BeginLine:      "start of line",
	EndLine:        "end of line",
	BeginText:      "start of text",
	EndText:        "end of text",
	WordBoundary:   "word boundary",
	NoWordBoundary: "no word boundary",
	AtLeast:        "at least",
	Times:          "times",
	To:             "to",
	From:           "from",
	Whitespace:     "whitespace",
}

// DefaultSyntaxConfig represents the default Syntax configuration
var DefaultSyntaxConfig = SyntaxConfig{
	Indent:   " ",
	UseComma: true,
}

func main() {
	testRegex := []string{
		"[^i*&2@]", // TODO Fix the problem that syntax.Regexp automatically 'inverts' this regex which leads to a massive description
		"DTB_[0-9]{3}_[A-Z]{3}",
		"A|B|C|D",
		"[A-Za-z]",
		"[A-Za-z _]",
		"[2-9]|[12]\\d|3[0-6]",
		"\\d+(\\.\\d\\d)?",
		"[ACE]",
		"[A-z]",
	}
	for _, reg := range testRegex {
		r, err := MakeRegex(reg, DescriptionRegexConfig{
			Syntax:      DefaultSyntaxConfig,
			Translation: GermanTranslationConfig,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(r.GetDescription())
	}

}

// MakeRegex creates a new DescriptionRegex with given regex and config.
func MakeRegex(regex string, config DescriptionRegexConfig) (*DescriptionRegex, error) {
	r, err := syntax.Parse(regex, syntax.Perl)
	if err != nil {
		return nil, err
	}

	return &DescriptionRegex{Regex: r, Config: config}, nil
}

// GetDescription returns the description of the DescriptionRegex r.
func (r *DescriptionRegex) GetDescription() string {
	return analyzeRegex(r.Regex, r.Config)
}

func analyzeRegex(reg *syntax.Regexp, config DescriptionRegexConfig) string {
	trans := config.Translation
	switch reg.Op {
	case syntax.OpEmptyMatch: // matches empty string
	case syntax.OpCharClass: // matches Runes interpreted as range pair list
		return getRuneDescription(reg.Rune, 2, trans.Or, config)
	case syntax.OpLiteral: // matches Runes sequence
		return getRuneDescription(reg.Rune, 1, "", config)
	case syntax.OpAnyCharNotNL: // matches any character except newline
		return fmt.Sprintf("%s %s \\n", trans.AnyChar, trans.Without)
	case syntax.OpAnyChar: // matches any character
		return trans.AnyChar
	case syntax.OpBeginLine: // matches empty string at beginning of line
		return trans.BeginLine
	case syntax.OpEndLine: // matches empty string at end of line
		return trans.EndLine
	case syntax.OpBeginText: // matches empty string at beginning of text
		return trans.BeginText
	case syntax.OpEndText: // matches empty string at end of text
		return trans.EndText
	case syntax.OpWordBoundary: // matches word boundary `\b`
		return trans.WordBoundary
	case syntax.OpNoWordBoundary: // matches word non-boundary `\B`
		return trans.NoWordBoundary
	case syntax.OpCapture: // capturing subexpression
		capture := make([]string, len(reg.Sub))
		for idx, sub := range reg.Sub {
			capture[idx] = analyzeRegex(sub, config)
		}
		return concat(capture, fmt.Sprintf(" %s ", trans.Or), config)
	case syntax.OpStar: // matches Sub[0] any amount of times
		return getRepeatDescription(0, -1, analyzeRegex(reg.Sub[0], config), config)
	case syntax.OpPlus: // matches at least one Sub[0]
		return getRepeatDescription(1, -1, analyzeRegex(reg.Sub[0], config), config)
	case syntax.OpQuest: // matches 0 or 1 time Sub[0]
		return getRepeatDescription(0, 1, analyzeRegex(reg.Sub[0], config), config)
	case syntax.OpRepeat: // matches Sub[0] between Min and Max times. When Max == -1 then Max should be infinite.
		return getRepeatDescription(reg.Min, reg.Max, analyzeRegex(reg.Sub[0], config), config)
	case syntax.OpConcat: // matches concatenation of Subs
		toConcat := make([]string, len(reg.Sub))
		for idx, sub := range reg.Sub {
			toConcat[idx] = analyzeRegex(sub, config)
		}
		return concat(toConcat, fmt.Sprintf(" %s ", trans.And), config)
	case syntax.OpAlternate: // matches alternation of Subs
		alternate := make([]string, len(reg.Sub))

		for idx, sub := range reg.Sub {
			alternate[idx] = analyzeRegex(sub, config)
		}
		return concat(alternate, fmt.Sprintf(" %s ", trans.Or), config)
	}
	return ""
}

func getRepeatDescription(min int, max int, char string, config DescriptionRegexConfig) string {
	trans := config.Translation
	switch {
	case min == max:
		return fmt.Sprintf("%d-%s %s", min, trans.Times, char)
	case max == -1:
		return fmt.Sprintf("%s %d-%s %s", trans.AtLeast, min, trans.Times, char)
	case min+1 == max:
		return fmt.Sprintf("%d %s %d-%s %s", min, trans.Or, max, trans.Times, char)
	default:
		return fmt.Sprintf("%s %d %s %d-%s %s", trans.Between, min, trans.And, max, trans.Times, char)
	}
}

func getRuneDescription(runes []rune, groupSize int, join string, config DescriptionRegexConfig) string {
	trans := config.Translation
	runeResult := make([]string, len(runes)/groupSize)
	for i, curRune := range runes {
		if i%groupSize != 0 {
			if curRune == runes[i-1] {
				continue
			}
			runeResult[i/groupSize] += fmt.Sprintf(" %s ", trans.And)
		} else {
			if isSameRune(runes[i : i+groupSize]) {
				runeResult[i/groupSize] += getStringBySymbol(curRune, trans)
				i += groupSize
				continue
			}
			runeResult[i/groupSize] += fmt.Sprintf("%s %s ", trans.AnyChar, trans.Between)
		}
		runeResult[i/groupSize] += getStringBySymbol(curRune, trans)
	}
	return fmt.Sprintf("(%s)", concat(runeResult, join, config))
}

func getStringBySymbol(r rune, trans TranslationConfig) string {
	reflectType := reflect.TypeOf(trans)
	reflectValue := reflect.ValueOf(trans)
	rString := strconv.Itoa(int(r))
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if tag, found := field.Tag.Lookup("special"); found && tag == rString {
			return reflectValue.Field(i).String()
		}
	}
	return string(r)
}

func isSameRune(runes []rune) bool {
	for _, r := range runes {
		if r != runes[0] {
			return false
		}
	}
	return true
}

func concat(toConcat []string, join string, config DescriptionRegexConfig) string {
	synt := config.Syntax
	if synt.UseComma && join != "" && len(toConcat) > 2 {
		return fmt.Sprintf("%s %s %s",
			strings.Join(toConcat[:len(toConcat)-1], ", "),
			join,
			toConcat[len(toConcat)-1],
		)
	}
	return strings.Join(toConcat, join)
}
