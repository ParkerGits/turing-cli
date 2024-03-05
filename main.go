package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

type TuringMachine struct {
	Start  string  `json:"start"`
	Accept string  `json:"accept"`
	Reject string  `json:"reject"`
	Delta  []Delta `json:"delta"`
}

type Delta struct {
	From string       `json:"from"`
	To   []Transition `json:"to"`
}

type Transition struct {
	Result []string `json:"result"`
	On     string   `json:"on"`
}

func main() {
	states := promptState()

	startResult := promptStart(states)

	acceptIdx, acceptResult := promptAccept(states)

	rejectResult := promptReject(states, acceptIdx)

	machine := TuringMachine{Start: startResult, Accept: acceptResult, Reject: rejectResult, Delta: make([]Delta, 0)}

	tapeAlphabet := promptTape()

	fromToMap := make(map[string][]Transition)
	for {
		optionsResult := promptOptions(fromToMap)
		// Finish.
		if strings.HasPrefix(optionsResult, "Finish") {
			break
		}

		// Remove a transition
		if strings.HasPrefix(optionsResult, "Remove") {
			promptRemove(fromToMap)
			continue
		}

		// Insert a transition
		fromIdx := promptFrom(states, acceptResult, rejectResult)
		toIdx := promptTo(states, acceptResult, rejectResult)

		onResult := promptOn(tapeAlphabet)

		writeResult := promptWrite(tapeAlphabet)

		directionResult := promptDirection()

		fromState := states[fromIdx]
		toState := states[toIdx]

		insertTransition(fromToMap, fromState, toState, onResult, writeResult, directionResult)
	}

	for from, to := range fromToMap {
		machine.Delta = append(machine.Delta, Delta{From: from, To: to})
	}

	outputResult := promptOutput()
	outfile, err := os.OpenFile(outputResult, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer outfile.Close()

	machineJson, err := json.Marshal(machine)
	if err != nil {
		panic("Error formatting Turing Machine as JSON.")
	}

	_, err = outfile.Write(machineJson)

	if err != nil {
		panic("Error writing the Turing Machine to the file.")
	}

	fmt.Printf("Successfully wrote Turing Machine to %s.\n", outputResult)
}

func promptState() []string {
	statesValidate := func(input string) error {
		number, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return errors.New("Invalid positive integer.")
		}
		if number <= 1 {
			return errors.New("Input must be >=2.")
		}
		return nil
	}

	statesPrompt := promptui.Prompt{
		Label:    "How many states in your Turing Machine?",
		Validate: statesValidate,
	}

	stateResult, err := statesPrompt.Run()
	if err != nil {
		panic("Failed to process states.")
	}

	nStates, err := strconv.ParseUint(stateResult, 10, 64)
	if err != nil {
		panic("Failed to parse input states.")
	}

	states := make([]string, nStates, nStates)
	for i := uint64(0); i < nStates; i++ {
		states[i] = strconv.FormatUint(i, 10)
	}

	return states
}

func promptStart(states []string) string {
	startPrompt := promptui.Select{
		Label: "Which state is your start state?",
		Items: states,
	}

	_, startResult, err := startPrompt.Run()
	if err != nil {
		panic("Failed to process start state.")
	}

	return startResult
}

func promptAccept(states []string) (int, string) {
	acceptPrompt := promptui.Select{
		Label: "Which state is your accept state?",
		Items: states,
	}

	acceptIdx, acceptResult, err := acceptPrompt.Run()
	if err != nil {
		panic("Failed to process accept state.")
	}

	return acceptIdx, acceptResult
}

func promptReject(states []string, acceptIdx int) string {
	stateStringsMinusAccept := removeIdx(states, acceptIdx)

	rejectPrompt := promptui.Select{
		Label: "Which state is your reject state?",
		Items: stateStringsMinusAccept,
	}

	_, rejectResult, err := rejectPrompt.Run()
	if err != nil {
		panic("Failed to process reject state.")
	}

	return rejectResult
}

func promptTape() []string {
	tapeValidate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Input should be nonempty.")
		}
		if strings.Contains(input, " ") {
			return errors.New("Input should contain no spaces.")
		}
		if strings.Contains(input, "_") {
			return errors.New("Input should contain no underscores.")
		}
		return nil
	}

	tapePrompt := promptui.Prompt{
		Label:    "What is the tape alphabet? Please enter a sequence of characters, not including '_'",
		Validate: tapeValidate,
	}

	tapeResult, err := tapePrompt.Run()
	if err != nil {
		panic("Failed to process reject state.")
	}

	return strings.Split(tapeResult, "")
}

func promptOptions(fromToMap map[string][]Transition) string {
	transitionStrings := make([]string, 0)
	for from, transitions := range fromToMap {
		transitionStrings = append(transitionStrings, transitionsToStrings(from, transitions)...)
	}
	for _, transition := range transitionStrings {
		fmt.Println(transition)
	}

	options := []string{"Add a transition."}
	if len(fromToMap) > 0 {
		options = append(options, "Remove a transition.")
	}
	options = append(options, "Finish.")

	optionsPrompt := promptui.Select{
		Label: "What would you like to do?",
		Items: options,
	}
	_, optionsResult, err := optionsPrompt.Run()
	if err != nil {
		panic("Failed to process selected option.")
	}

	return optionsResult
}

func promptFrom(states []string, accept, reject string) int {
	fromPrompt := promptui.Select{
		Label: "From which state?",
		Items: statesWithAcceptReject(states, accept, reject),
	}

	fromIdx, _, err := fromPrompt.Run()
	if err != nil {
		panic("Failed to process 'from' state.")
	}

	return fromIdx
}

func promptTo(states []string, accept, reject string) int {
	toPrompt := promptui.Select{
		Label: "To which state?",
		Items: statesWithAcceptReject(states, accept, reject),
	}

	toIdx, _, err := toPrompt.Run()
	if err != nil {
		panic("Failed to process 'to' state.")
	}

	return toIdx
}

func promptOn(tapeAlphabet []string) string {
	onPrompt := promptui.Select{
		Label: "On what input? '_' for blank",
		Items: append(tapeAlphabet, "_"),
	}

	_, onResult, err := onPrompt.Run()
	if err != nil {
		panic("Failed to process input symbol.")
	}

	return onResult
}

func promptWrite(tapeAlphabet []string) string {
	writePrompt := promptui.Select{
		Label: "What symbol does the head write? '_' for blank",
		Items: append(tapeAlphabet, "_"),
	}

	_, writeResult, err := writePrompt.Run()
	if err != nil {
		panic("Failed to process write symbol.")
	}

	return writeResult
}

func promptDirection() string {
	directionPrompt := promptui.Select{
		Label: "What direction does the head move?",
		Items: []string{"Left", "Right"},
	}
	directionIdx, _, err := directionPrompt.Run()
	if err != nil {
		panic("Failed to process direction.")
	}

	if directionIdx == 0 {
		return "L"
	}
	return "R"
}

func insertTransition(fromToMap map[string][]Transition, from, to, on, write, direction string) {
	_, ok := fromToMap[from]

	if !ok {
		fromToMap[from] = []Transition{}
	}

	existingTransitions, _ := fromToMap[from]

	for _, transition := range existingTransitions {
		if transition.On == on {
			fmt.Printf("Duplicate transition from state %s on input %s.\n", from, on)
			return
		}
	}

	fromToMap[from] = append(existingTransitions, Transition{Result: []string{to, write, direction}, On: on})
}

func promptOutput() string {
	outputValidate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Expecting nonempty string.")
		}
		return nil
	}
	outputPrompt := promptui.Prompt{
		Label:    "Which file would you like to write your new Turing Machine to?",
		Validate: outputValidate,
	}

	outputResult, err := outputPrompt.Run()
	if err != nil {
		panic("Failed to process output file name.")
	}

	return outputResult
}

func promptRemove(fromToMap map[string][]Transition) string {
	states := make([]string, 0)
	for state := range fromToMap {
		states = append(states, state)
	}

	removeStatePrompt := promptui.Select{
		Label: "Which state would you like to remove a transition from?",
		Items: states,
	}
	_, removeStateResult, err := removeStatePrompt.Run()
	if err != nil {
		panic("Failed to process transition state.")
	}

	transitions := fromToMap[removeStateResult]
	fmt.Println(transitionsToStrings(removeStateResult, transitions))
	removeTransitionPrompt := promptui.Select{
		Label: "Which transition would you like to remove?",
		Items: transitionsToStrings(removeStateResult, transitions),
	}
	removeTransitionIdx, removeTransitionResult, err := removeTransitionPrompt.Run()
	if err != nil {
		panic("Failed to process transition state.")
	}

	newTransitions := removeIdx(fromToMap[removeStateResult], removeTransitionIdx)
	fromToMap[removeStateResult] = newTransitions
	return removeTransitionResult
}

func transitionsToStrings(from string, transitions []Transition) []string {
	ret := make([]string, 0)
	for _, transition := range transitions {
		ret = append(ret, fmt.Sprintf("∂(%s, %s) = (%s, %s, %s)", from, transition.On, transition.Result[0], transition.Result[1], transition.Result[2]))
	}
	return ret
}

func statesWithAcceptReject(states []string, accept, reject string) []string {
	ret := make([]string, len(states))

	for i, state := range states {
		if state == accept {
			ret[i] = fmt.Sprintf("%s (accept)", state)
			continue
		}
		if state == reject {
			ret[i] = fmt.Sprintf("%s (reject)", state)
			continue
		}
		ret[i] = state
	}

	return ret
}

func removeIdx[T any](s []T, idx int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:idx]...)
	return append(ret, s[idx+1:]...)
}
