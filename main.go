package main

import (
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"strings"
	"path/filepath"
	"github.com/z7zmey/php-parser/php7"
	"bytes"
	"pilot114/ctp/walkers"
	"pilot114/ctp/structure"
	"bufio"
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

// инициализируем граф и корневую ноду
var graph structure.FindInfoGraph
var rootNode structure.Node

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

func findInFile(parent structure.Node, path string, signature walkers.Signature) {

	// парсинг php файла в AST
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
	}
	src := bytes.NewBufferString(string(file))
	parser := php7.NewParser(src, path)
	parser.Parse()
	for _, e := range parser.GetErrors() {
		fmt.Println(e)
		os.Exit(0)
	}

	// сначала проходим по файлу простым поиском
	find := strings.Join([]string{"->", signature.MethodName, "("}, "")

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 1
	findInfos := []walkers.FindInfo{}

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), find) {
			findInfos = append(findInfos, walkers.FindInfo{
				Line: scanner.Text(),
				LineNumber: line,
				Signature: signature,
				Node: nil,
			})
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		return
	}

	// проходим по AST, чтобы получить нужные ноды
	for _, findInfo := range findInfos {
		nodeFinder := walkers.NodeFinder{
			Writer:        os.Stdout,
			Positions:     parser.GetPositions(),
			FindSignature: signature.MethodName,
			FindInfo:      &findInfo,
		}

		parser.GetRootNode().Walk(nodeFinder)

		// добавляем найденное в граф
		n := structure.Node{Value: findInfo.Node}
		graph.AddNode(&n)
		graph.AddEdge(&n, &parent)
	}
}

// TODO: читать один раз и затем искать в памяти
func find(parent structure.Node, root string, dirs []ScanDir, signature walkers.Signature) {

	for _, dir := range dirs {
		var path = strings.Join([]string{root, dir.Name}, "/")
		fmt.Println("Dir:", path)

		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".php") {
				findInFile(parent, path, signature)
			}
			return nil
		})
	}
}

func main() {
	conf := loadConfig()
	phpShtormReference := "\\App\\Model\\PO\\PayItem::removeByOrderIdXXX"

	parts := strings.Split(phpShtormReference, "::")
	signature := walkers.Signature{
		Namespace: parts[0],
		MethodName: parts[1],
	}

	rootNode = structure.Node{
		Value: nil,
	}
	graph.AddNode(&rootNode)

	find(rootNode, conf.RootDir, conf.ScanDirs, signature)
	//fmt.Println("Count finded:", len(finded))

	graph.String()
}