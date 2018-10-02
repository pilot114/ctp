package main

import (
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"strings"
	"path/filepath"
	"bufio"
)

type Config struct {
	RootDir string
	ScanDirs []string
	CountWorkers int
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

	fmt.Println(path)
	for scanner.Scan() {
		// TODO WTF
		fmt.Println(line)

		if strings.Contains(scanner.Text(), search) {
			fmt.Println(scanner.Text(), path, line)

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

//"App" , "www", "cron", "html", "lib"
func main() {
	var conf Config = loadConfig()

	for _, item := range conf.ScanDirs {
		var path = strings.Join([]string{conf.RootDir, item}, "/")
		fmt.Println("Dir:", path)

		finded := []FindInfo{}
		findedInFile := []FindInfo{}

		// TODO: читать один раз и затем искать в памяти
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".php") {
				findedInFile = findInFile(path, "_formatData")
				finded = append(finded, findedInFile...)
			}
			return nil
		})
		fmt.Println("Count finded:", len(finded))
	}
}