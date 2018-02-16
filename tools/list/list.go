package main

import (
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

func main() {
	files, err := getTestFiles()
	if err != nil {
		panic(err)
	}
	tests := getTests(files)

	for _, t := range tests {
		fmt.Println(formatCommand(t))
	}
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
					i := strings.Index(tf, filepath.Join("github.com", "Azure-Samples", "azure-sdk-for-go-samples"))
					pack, _ := filepath.Split(tf[i:])

					t := test{
						example: f.Name.Name,
						pack:    pack,
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
	return filepath.Join(gopath, "src", "github.com", "Azure-Samples", "azure-sdk-for-go-samples")
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
