package main

import (
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"bufio"
	"strings"
	"path/filepath"
	"github.com/z7zmey/php-parser/php7"
	"bytes"
)

type Config struct {
	RootDir string
	ScanDirs []ScanDir
	CountWorkers int
}

type ScanDir struct {
	Name string
	Namespace string
}

type FindInfo struct {
	FileName string
	Line string
	LineNumber int
}


func loadConfig() Config {
	configContent, e := ioutil.ReadFile("./config.json")

	if e != nil {
		fmt.Printf("Read config error: %v\n", e)
		os.Exit(1)
	}

	var conf Config
	json.Unmarshal(configContent, &conf)
	return conf
}

func findInFile(path string, search string) []FindInfo {
	f, err := os.Open(path)
	if err != nil {
		return []FindInfo{} // TODO
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 1
	finded := []FindInfo{}

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), search) {
			finded = append(finded, FindInfo{
				FileName: path,
				Line: scanner.Text(),
				LineNumber: line,
			})
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		return []FindInfo{} // TODO
	}
	return finded
}

// TODO: читать один раз и затем искать в памяти
func findInDir(path string, search string) []FindInfo {
	finded := []FindInfo{}
	findedInFile := []FindInfo{}
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".php") {
			findedInFile = findInFile(path, search)
			finded = append(finded, findedInFile...)
		}
		return nil
	})
	return finded
}

func main() {
	conf := loadConfig()
	phpShtormReference := "\\App\\Model\\PO\\PayItem::removeByOrderIdXXX"
	//\company_payorder::items_delete

	parts := strings.Split(phpShtormReference, "::")
	namespace, method := parts[0], parts[1]
	fmt.Println(namespace)

	finded := []FindInfo{}

	for _, dir := range conf.ScanDirs {
		var path = strings.Join([]string{conf.RootDir, dir.Name}, "/")
		fmt.Println("Dir:", path)
		finded = findInDir(path, strings.Join([]string{"->", method, "("}, ""))

		fmt.Println("Count finded:", len(finded))
	}

	for _, findInfo := range finded {
		fmt.Println(findInfo.FileName)
		fmt.Println(findInfo.LineNumber)
		fmt.Println(findInfo.Line)

		file, err := ioutil.ReadFile(findInfo.FileName)
		if err != nil {
			fmt.Print(err)
		}

		// парсинг php файла
		src := bytes.NewBufferString(string(file))
		parser := php7.NewParser(src, findInfo.FileName)
		parser.Parse()

		for _, e := range parser.GetErrors() {
			fmt.Println(e)
		}

		visitor := Walker{
			Writer:    os.Stdout,
			Positions: parser.GetPositions(),
		}

		rootNode := parser.GetRootNode()
		rootNode.Walk(visitor)

		os.Exit(0)
	}
}