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
	r, _ := MakeRegex("[A-Z]|ALE|[0-9]")

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

	visited := make([]bool, len(program.Inst))
	desc := analyzeInstructions(program.Inst, idx, end, &visited)

	return desc
}

func analyzeInstructions(instruction []syntax.Inst, idx int, end int, visited *[]bool) string {
	if (*visited)[idx] {
		return ""
	}
	result := []string{""}
	newGroup := false
	for idx != end {
		instr := instruction[idx]

		if (*visited)[idx] {
			break
		}
		(*visited)[idx] = true

		if newGroup {
			newGroup = false
			result = append(result, "")
		}

		resultIdx := len(result) - 1

		switch instr.Op {
		case syntax.InstAlt:
			result[resultIdx] += analyzeInstructions(instruction, int(instr.Out), end, visited) + " ODER "
			idx = int(instr.Arg)
			continue
		case syntax.InstAltMatch:
		case syntax.InstCapture:
		case syntax.InstEmptyWidth:
		case syntax.InstMatch:
		case syntax.InstFail:
		case syntax.InstNop:
		case syntax.InstRune:
			runeResult := make([]string, len(instr.Rune)/2)
			for i, curRune := range instr.Rune {
				if i%2 != 0 {
					continue
				}
				runeResult[i/2] = string(curRune)
				if i+1 < len(instr.Rune) && curRune != instr.Rune[i+1] {
					runeResult[i/2] += fmt.Sprintf(" BIS %s", string(instr.Rune[i+1]))
				}
			}
			result[resultIdx] += strings.Join(runeResult, " ODER ")
		case syntax.InstRune1:
			result[len(result)-1] += string(instr.Rune[0])
		case syntax.InstRuneAny:
		case syntax.InstRuneAnyNotNL:
		}

		// result = append(result, instr.Op.String())

		nextIdx := int(instr.Out)

		idx = nextIdx
	}
	return strings.Join(result, "")
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
