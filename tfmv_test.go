package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"
)

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func moveFile(filename, srcDir, destDir string) error {
	src, err := filepath.Abs(srcDir + filename)
	if err != nil {
		return err
	}
	dest, err := filepath.Abs(destDir + filename)
	if err != nil {
		return err
	}
	return os.Rename(src, dest)
}

func removeFile(filename string) error {
	path, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	if fileExists(path) {
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	return nil
}

// https://github.com/palantir/tfjson/blob/57123411e29c8945cd8dc89b6237c8f6f31ddf6e/tfjson_test.go#L124-L132
func mustRun(t *testing.T, name string, arg ...string) {
	if _, err := exec.Command(name, arg...).Output(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			t.Fatal(string(exitError.Stderr))
		} else {
			t.Fatal(err)
		}
	}
}

func testModulePath(t *testing.T, modulePath string) string {
	module, err := filepath.Abs(modulePath)
	assert.NoError(t, err)
	return module
}

func initModule(t *testing.T, modulePath string) {
	module := testModulePath(t, modulePath)
	mustRun(t, "terraform", "init", module)
}

func generatePlan(t *testing.T, modulePath string) string {
	planFile, err := ioutil.TempFile("", "terraform-plan")
	assert.NoError(t, err)
	err = planFile.Close()
	assert.NoError(t, err)

	module := testModulePath(t, modulePath)

	planPath := planFile.Name()
	mustRun(t, "terraform", "plan", "-out="+planPath, module)
	return planPath
}

func getTestPlan(t *testing.T, modulePath string) *terraform.Plan {
	initModule(t, modulePath)
	planPath := generatePlan(t, modulePath)
	plan, err := getPlan(planPath)
	assert.NoError(t, err)
	return plan
}

func TestMain(m *testing.M) {
	// clean up the generated state file before and after the test run
	stateFile := "terraform.tfstate"

	err := removeFile(stateFile)
	if err != nil {
		log.Fatalln(err)
	}

	exitCode := m.Run()

	err = removeFile(stateFile)
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(exitCode)
}

func TestMissingPlan(t *testing.T) {
	_, err := getPlan("missing-file")
	assert.Error(t, err)
}

func TestEmptyPlan(t *testing.T) {
	plan := getTestPlan(t, "test/empty")
	assert.True(t, plan.Diff.Empty())
	assert.Len(t, plan.Diff.Modules, 1)
}

func TestSimplePlan(t *testing.T) {
	plan := getTestPlan(t, "test/simple")
	assert.False(t, plan.Diff.Empty())
	assert.Len(t, plan.Diff.Modules, 1)
}

func TestChangesByType_Simple(t *testing.T) {
	plan := getTestPlan(t, "test/simple")

	changesByType, err := getChangesByType(plan)
	assert.NoError(t, err)

	types := changesByType.GetTypes()
	assert.Equal(t, []ResourceType{"local_file"}, types)

	changes := changesByType.Get("local_file")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 0)
}

func TestChangesByType_Multi(t *testing.T) {
	plan := getTestPlan(t, "test/multi")

	changesByType, err := getChangesByType(plan)
	assert.NoError(t, err)

	types := changesByType.GetTypes()
	assert.Len(t, types, 2)

	changes := changesByType.Get("local_file")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 0)

	changes = changesByType.Get("tls_private_key")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 0)
}

func TestChangesAfterApplyAndMove(t *testing.T) {
	rootModule := "test/module_ref/"
	mustRun(t, "terraform", "apply", "-auto-approve", rootModule)

	filename := "tls.tf"
	destModule := "test/empty/"
	err := moveFile(filename, rootModule, destModule)
	assert.NoError(t, err)

	// move the file back
	defer func() {
		err = moveFile(filename, destModule, rootModule)
		assert.NoError(t, err)
	}()

	plan := getTestPlan(t, rootModule)
	changesByType, err := getChangesByType(plan)
	assert.NoError(t, err)

	types := changesByType.GetTypes()
	assert.Equal(t, []ResourceType{"tls_private_key"}, types)

	changes := changesByType.Get("tls_private_key")
	assert.Len(t, changes.Created, 1)
	assert.Len(t, changes.Destroyed, 1)
}
