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

	visited := make([]bool, len(program.Inst))
	result := analyzeInstructions(program.Inst, idx, end, &visited)
	desc := strings.Join(result, " - ")

	return desc
}

func analyzeInstructions(instruction []syntax.Inst, origin int, end int, visited *[]bool) []string {
	if (*visited)[origin] {
		return []string{""}
	}
	result := make([]string, 1)
	newGroup := false
	idx := origin
	start := true
	for idx != len(instruction) && idx != end {
		instr := instruction[idx]

		if !start && (idx == origin || (*visited)[idx]) {
			idx++
			continue
		}
		(*visited)[idx] = true
		start = false

		if newGroup {
			newGroup = false
			result = append(result, "")
		}

		resultIdx := len(result) - 1

		switch instr.Op {
		case syntax.InstAlt:
			result[resultIdx] += strings.Join(analyzeInstructions(instruction, int(instr.Out), end, visited), " ODER ")
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

		if nextIdx == end {
			idx++
			newGroup = true
			continue
		}

		idx = nextIdx
	}
	for i, r := range result {
		result[i] = fmt.Sprintf("(%s)", r)
	}
	return result
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
