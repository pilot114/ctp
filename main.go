package main

import (
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"bufio"
	"strings"
	"path/filepath"
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
	var conf = loadConfig()
	var phpShtormReference = "\\App\\Model\\PO\\PayItem::removeByOrderId"
	//\company_payorder::items_delete

	var parts = strings.Split(phpShtormReference, "::")
	var namespace, method = parts[0], parts[1]
	fmt.Println(namespace)

	for _, dir := range conf.ScanDirs {
		var path = strings.Join([]string{conf.RootDir, dir.Name}, "/")
		fmt.Println("Dir:", path)
		finded := findInDir(path, strings.Join([]string{"->", method, "("}, ""))

		fmt.Println("Count finded:", len(finded))
		//fmt.Println("Finded:", finded)
	}
}