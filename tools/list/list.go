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

	b, err := json.MarshalIndent(tasks, "", "    ")
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
			Name: fmt.Sprintf("%s/%s", filepath.Base(t.pack), strings.TrimPrefix(t.example, "Example")),
			Settings: a01TaskSetting{
				Execution: map[string]string{
					"command": formatCommand(t),
				},
			},
		})
	}
	return tasks
}

type a01TaskSetting struct {
	Version     string            `json:"ver,omitempty"`
	Execution   map[string]string `json:"execution,omitempty"`
	Classifier  map[string]string `json:"classifier,omitempty"`
	Miscellanea map[string]string `json:"msic,omitempty"`
}

type a01Task struct {
	Annotation    string                 `json:"annotation,omitempty"`
	Duration      int                    `json:"duration,omitempty"`
	ID            int                    `json:"id,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Result        string                 `json:"result,omitempty"`
	ResultDetails map[string]interface{} `json:"result_details,omitempty"`
	RunID         int                    `json:"run_id,omitempty"`
	Settings      a01TaskSetting         `json:"settings,omitempty"`
	Status        string                 `json:"status,omitempty"`
}
