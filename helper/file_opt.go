package helper

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func DirIsEmpty(dirPath string) (bool, error) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	return len(dirEntries) == 0, err
}

func FunctionExists(dirPath, funcName string) (bool, error) {
	isEmpty, err := DirIsEmpty(dirPath)
	if err != nil {
		return false, err
	}
	if isEmpty {
		return false, nil
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			if exists, err := parseGoFileForFunction(path, funcName); err == nil && exists {
				return fmt.Errorf("found:%s", path)
			}
		}
		return nil
	})

	if err != nil && strings.HasPrefix(err.Error(), "found:") {
		return true, nil
	}

	return false, err
}

func parseGoFileForFunction(filePath, funcName string) (bool, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, filePath, src, parser.AllErrors)
	if err != nil {
		return false, err
	}

	funcMap := make(map[string]bool)
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			funcMap[fn.Name.Name] = true
		}
		return true
	})

	return funcMap[funcName], nil
}
