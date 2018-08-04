package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"flag"
	tfmt "github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/terraform"
)

const MatchModeFirstMatching = "first-matching"
const MatchModeSameName = "same-name"

func getPlan(file string) (fmtPlan *tfmt.Plan, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
	}()

	// Terraform has two Plan types, for some reason. `terraform.Plan` doesn't include the module in the address, so only use it for reading from the plan file, then convert it to the other one.
	plan, err := terraform.ReadPlan(f)
	if err != nil {
		return nil, err
	}
	fmtPlan = tfmt.NewPlan(plan)
	return fmtPlan, nil
}

func getChangesByType(plan *tfmt.Plan) (ChangesByType, error) {
	changesByType := ChangesByType{}

	// inspired by
	// https://github.com/palantir/tfjson/blob/master/tfjson.go
	for _, resource := range plan.Resources {
		changesByType.Add(*resource)
	}

	return changesByType, nil
}

func checkIfObjectsMatch(name string, creation, deletion interface{}) error {
	if reflect.DeepEqual(creation, deletion) {
		err := fmt.Errorf(name+" match, which they shouldn't:\ncreation: %+v\ndeletion:%+v\n", creation, deletion)
		return err
	}
	return nil
}

func matchSameName(creation tfmt.InstanceDiff, changes *ResourceChanges) *tfmt.InstanceDiff {
	for _, destruction := range changes.Destroyed {
		if creation.Addr.Name == destruction.Addr.Name {
			return &destruction

		}
	}
	return nil
}

func appendMove(moves []string, creation, deletion tfmt.InstanceDiff) ([]string, error) {
	// sanity checks
	if err := checkIfObjectsMatch("Addrs", creation.Addr, deletion.Addr); err != nil {
		return moves, err
	}
	if err := checkIfObjectsMatch("InstanceDiffs", creation, deletion); err != nil {
		return moves, err
	}

	moves = append(moves, "terraform state mv "+deletion.Addr.String()+" "+creation.Addr.String())
	return moves, nil
}

func getMoveStatementsForFirstMatching(plan *tfmt.Plan) ([]string, error) {
	moves := []string{}

	changesByType, err := getChangesByType(plan)
	if err != nil {
		return moves, err
	}

	for _, changes := range changesByType {
		for i, creation := range changes.Created {
			// stop if we're out of matches
			if i == len(changes.Destroyed) {
				break
			}
			deletion := changes.Destroyed[i]

			moves, err = appendMove(moves, creation, deletion)
			if err != nil {
				return moves, err
			}
		}
	}

	return moves, nil
}

func getMoveStatementsForSameName(plan *tfmt.Plan) ([]string, error) {
	moves := []string{}

	changesByType, err := getChangesByType(plan)
	if err != nil {
		return moves, err
	}

	for _, changes := range changesByType {
		for _, creation := range changes.Created {

			match := matchSameName(creation, changes)
			if match != nil {
				if moves, err = appendMove(moves, creation, *match); err != nil {
					return moves, err
				}

			}
		}
	}

	return moves, nil
}

func getMoveStatements(plan *tfmt.Plan, mode string) ([]string, error) {
	switch mode {
	case MatchModeFirstMatching:
		return getMoveStatementsForFirstMatching(plan)
	case MatchModeSameName:
		return getMoveStatementsForSameName(plan)
	default:
		return nil, fmt.Errorf("unknown match-mode %s, available modes are %s (default) and %s", mode, MatchModeFirstMatching, MatchModeSameName)
	}
}

func main() {
	// TODO parameterize
	mode := flag.String("mode", "first-matching", "mode to use when matching resources to each other. Can be first-matching or same-name")

	flag.Parse()
	planfile := "tfplan"
	plan, err := getPlan(planfile)
	if err != nil {
		log.Fatalln(err)
	}

	moves, err := getMoveStatements(plan, *mode)
	if err != nil {
		log.Fatalln(err)
	}
	for _, move := range moves {
		fmt.Println(move)
	}
}
