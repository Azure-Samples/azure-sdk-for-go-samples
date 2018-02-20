package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	repo = filepath.Join("github.com", "Azure-Samples", "azure-sdk-for-go-samples")
)

func main() {
	files, err := getTestFiles()
	if err != nil {
		panic(err)
	}
	tests := getTests(files)
	tasks := convertToTasks(tests)

	b, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	/*
		execution := strings.Fields(tasks[len(tasks)-1].Settings.Execution["command"])
		var cmd *exec.Cmd
		if len(execution) < 2 {
			cmd = exec.Command(execution[0])
		} else {
			cmd = exec.Command(execution[0], execution[1:]...)
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(output))
	*/
}

type test struct {
	example string
	pack    string
}

func formatCommand(t test) string {
	return fmt.Sprintf("go test -v %s -run ^%s$", t.pack, t.example)
}

func getTests(testFiles []string) []test {
	tests := []test{}
	for _, tf := range testFiles {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, tf, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		for _, d := range node.Decls {
			if f, ok := d.(*ast.FuncDecl); ok {
				if strings.HasPrefix(f.Name.Name, "Example") {
					i := strings.Index(tf, repo)
					pack, _ := filepath.Split(tf[i:])

					t := test{
						example: f.Name.Name,
						pack:    filepath.Clean(pack),
					}

					tests = append(tests, t)
				}
			}
		}
	}
	return tests
}

func getRoot() string {
	gopath := build.Default.GOPATH
	return filepath.Join(gopath, "src", repo)
}

func getTestFiles() ([]string, error) {
	testFiles := []string{}
	rootDir := getRoot()
	vendorDir := filepath.Join(rootDir, "vendor")
	err := filepath.Walk(rootDir, func(path string, f os.FileInfo, err error) error {
		if !strings.HasPrefix(path, vendorDir) {
			if strings.HasSuffix(path, "_test.go") {
				testFiles = append(testFiles, path)
			}
		}
		return nil
	})
	return testFiles, err
}

func convertToTasks(tests []test) []a01Task {
	tasks := []a01Task{}
	for _, t := range tests {
		tasks = append(tasks, a01Task{
			Version: "1.0",
			Execution: a01TaskExecution{
				Command: formatCommand(t),
			},
			Classifier: a01TaskClassifier{
				Identifier: fmt.Sprintf("%s/%s", strings.TrimPrefix(t.pack, repo+string(filepath.Separator)), strings.TrimPrefix(t.example, "Example")),		
				Type:       Live,
			},
		})
	}
	return tasks
}

type testType string

const (
	Recording testType = "Recording"
	Live testType = "Live"
	Unit testType = "Unit"
)

type a01TaskExecution struct {
	Command string `json:"command,omitempty"`
}

type a01TaskClassifier struct {
	Identifier string `json:"identifier,omitempty"`
	Type       testType `json:"type,omitempty"`
}

type a01Task struct {
	Version    string            `json:"ver,omitempty"`
	Execution  a01TaskExecution  `json:"execution,omitempty"`
	Classifier a01TaskClassifier `json:"classifier,omiyempty"`
}
