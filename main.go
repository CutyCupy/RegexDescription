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
	r, _ := MakeRegex("[a-zA-Z]|AL|EX")

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
	program, err := syntax.Compile(r.Regex)
	if err != nil {
		return "Fehlerhafter Regex"
	}

	idx := program.Start

	end := len(program.Inst) - 1

	result := analyzeInstructions(program.Inst, idx, end)
	desc := strings.Join(result, " - ")

	return desc
}

func analyzeInstructions(instruction []syntax.Inst, idx int, end int) []string {
	result := make([]string, 1)
	newGroup := false
	for idx != len(instruction) && idx != end {
		instr := instruction[idx]

		switch instr.Op {
		case syntax.InstAlt:
			result := analyzeInstructions(instruction[:idx], int(instr.Out), end)
			return []string{strings.Join(result, " ODER ")}
		case syntax.InstAltMatch:
		case syntax.InstCapture:
		case syntax.InstEmptyWidth:
		case syntax.InstMatch:
		case syntax.InstFail:
		case syntax.InstNop:
		case syntax.InstRune:

		case syntax.InstRune1:
			result = append(result, string(instr.Rune[0]))
		case syntax.InstRuneAny:
		case syntax.InstRuneAnyNotNL:
		}

		// result = append(result, instr.Op.String())

		nextIdx := int(instr.Out)

		if nextIdx == end {
			idx++
			continue
		}

		idx = nextIdx
	}

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
