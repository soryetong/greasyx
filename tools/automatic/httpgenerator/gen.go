package httpgenerator

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/tools/automatic/config"
)

type XContext struct {
	ModuleName     string
	Src            string
	Output         string
	RouterPrefix   string
	NeedRequestLog bool

	TypesPackageName string
	TypesPackagePath string
	LogicPackagePath map[string]string
	LogicFuncName    map[string]string
	LogicPackageName map[string]string
	LogicName        map[string]string
	HandlerPackPath  map[string]string
	HandlerPackName  map[string]string
	RouterPath       string

	FileType config.FileType
	Types    []*TypesStructSpec
	Services []*ServiceSpec
}

func (self *HttpGenerator) Generate() (err error) {
	info, err := os.Stat(self.Src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		err = filepath.WalkDir(self.Src, func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() == false {
				err = self.start(filePath)
				if err != nil {
					return err
				}
			}

			return nil
		})

		return err
	}

	return self.start(self.Src)
}

func (self *HttpGenerator) start(filename string) (err error) {
	console.Echo.Debugf("开始API文件: %s 内容读取", filename)
	fileContentByte, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成API文件解析\n")

	filenameArr := strings.Split(filepath.Base(filename), ".")
	if len(filenameArr) <= 0 {
		return errors.New("文件名不合法")
	}
	fileContent := string(fileContentByte)
	groupName := filenameArr[0]

	// Parse the structs and routes.
	console.Echo.Debug("开始Struct内容解析")
	if err = self.PTypesStruct(fileContent); err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成Struct内容解析\n")

	console.Echo.Debug("开始Service服务解析")
	if err = self.PRoutesService(fileContent); err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成Service服务解析\n")

	// Generate types.
	console.Echo.Debug("开始Struct代码生成")
	if err = self.GenTypes(); err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成Struct代码生成\n")

	// Generate logic.
	console.Echo.Debug("开始Logic代码生成")
	if err = self.GenLogic(); err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成Logic代码生成\n")

	// Generate handler.
	console.Echo.Debug("开始Handler代码生成")
	if err = self.GenHandler(); err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成Handler代码生成\n")

	// Generate router.
	console.Echo.Debug("开始Router代码生成")
	if err = self.GenRouter(groupName); err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成Router代码生成\n")

	// Generate server.
	console.Echo.Debug("开始Server代码生成")
	if err = self.GenServer(); err != nil {
		return err
	}
	console.Echo.Debug("✅ 已完成Server代码生成\n")
	console.Echo.Infof("ℹ️ 提示: 文件: %s 代码生成已完成\n", filename)

	return nil
}

// normalFormatFileWithGofmt formats a file using gofmt.
func (self *HttpGenerator) normalFormatFileWithGofmt(filepath string) {
	cmd := exec.Command("gofmt", filepath)
	if err := cmd.Run(); err != nil {
		return
	}
}

// formatFileWithGofmt formats a file using gofmt.
func (self *HttpGenerator) formatFileWithGofmt(filepath string) {
	originalContent, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filepath, err)
		return
	}

	// Run gofmt on the original content.
	cmd := exec.Command("gofmt")
	cmd.Stdin = bytes.NewReader(originalContent)
	var formattedContent bytes.Buffer
	cmd.Stdout = &formattedContent
	if err = cmd.Run(); err != nil {
		fmt.Printf("Error running gofmt on file %s: %v\n", filepath, err)
		return
	}

	// Split lines and trim trailing empty lines
	formattedLines := bytes.Split(formattedContent.Bytes(), []byte("\n"))
	trimmedFormattedLines := trimTrailingEmptyLines(formattedLines)

	// Reassemble the final content
	var finalContent bytes.Buffer
	for i, line := range trimmedFormattedLines {
		finalContent.Write(line)
		if i < len(trimmedFormattedLines)-1 || len(line) > 0 { // Add newline except for the last empty line
			finalContent.WriteString("\n")
		}
	}

	// Write the formatted content back to the file.
	if err = os.WriteFile(filepath, finalContent.Bytes(), 0644); err != nil {
		fmt.Printf("Error writing formatted content to file %s: %v\n", filepath, err)
	}
}

// trimTrailingEmptyLines removes trailing empty lines from a slice of lines.
func trimTrailingEmptyLines(lines [][]byte) [][]byte {
	end := len(lines)
	for end > 0 && len(bytes.TrimSpace(lines[end-1])) == 0 {
		end--
	}
	return lines[:end]
}
