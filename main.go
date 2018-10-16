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
	"pilot114/ctp/walkers"
	"pilot114/ctp/structure"
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

func findInFile(path string, signature walkers.Signature) []walkers.FindInfo {

	find := strings.Join([]string{"->", signature.MethodName, "("}, "")

	f, err := os.Open(path)
	if err != nil {
		return []walkers.FindInfo{} // TODO error handle
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 1
	finded := []walkers.FindInfo{}

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), find) {
			finded = append(finded, walkers.FindInfo{
				FileName: path,
				Line: scanner.Text(),
				LineNumber: line,
				Signature: signature,
			})
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		return []walkers.FindInfo{} // TODO error handle
	}
	return finded
}

// TODO: читать один раз и затем искать в памяти
func findInDir(path string, signature walkers.Signature) []walkers.FindInfo {

	finded := []walkers.FindInfo{}
	findedInFile := []walkers.FindInfo{}
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".php") {
			findedInFile = findInFile(path, signature)
			finded = append(finded, findedInFile...)
		}
		return nil
	})

	return finded
}

// инициализируем граф
var graph structure.FindInfoGraph

// итерация заполнения графа
func addFindedToGraph(signature walkers.Signature, parent structure.Node, fileName string, findInfos []walkers.FindInfo) {

	// парсинг php файла в AST
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Print(err)
	}
	src := bytes.NewBufferString(string(file))
	parser := php7.NewParser(src, fileName)
	parser.Parse()
	for _, e := range parser.GetErrors() {
		fmt.Println(e)
		os.Exit(0)
	}

	// проходим по AST, чтобы получить нужные ноды
	for _, findInfo := range findInfos {

		nodeFinder := walkers.NodeFinder{
			Writer:        os.Stdout,
			Positions:     parser.GetPositions(),
			FindSignature: signature.MethodName,
			FindInfo:      &findInfo,
		}

		rootNode := parser.GetRootNode()
		rootNode.Walk(nodeFinder)

		// добавляем найденное в граф
		node := structure.Node{findInfo}
		graph.AddNode(&node)
		graph.AddEdge(&node, &parent)
	}
}

func find(root string, dirs []ScanDir, signature walkers.Signature) []walkers.FindInfo {

	finded := []walkers.FindInfo{}
	for _, dir := range dirs {
		var path = strings.Join([]string{root, dir.Name}, "/")
		fmt.Println("Dir:", path)
		finded = findInDir(path, signature)
	}
	return finded
}

func main() {
	conf := loadConfig()
	phpShtormReference := "\\App\\Model\\PO\\PayItem::removeByOrderIdXXX"

	parts := strings.Split(phpShtormReference, "::")
	signature := walkers.Signature{
		Namespace: parts[0],
		MethodName: parts[1],
	}

	rootNode := structure.Node{
		Value: walkers.FindInfo{
			FileName: "ROOT",
			Line: "ROOT",
			LineNumber: 0,
			Node: nil,
		},
	}
	graph.AddNode(&rootNode)

	finded := find(conf.RootDir, conf.ScanDirs, signature)
	fmt.Println("Count finded:", len(finded))

	// делаем мапу, группируя найденное по именам файлов
	FindInfoMap := make(map[string][]walkers.FindInfo)
	for _, findInfo := range finded {
		FindInfoMap[findInfo.FileName] = append(FindInfoMap[findInfo.FileName], findInfo)
	}

	for fileName, findInfos := range FindInfoMap {
		addFindedToGraph(signature, rootNode, fileName, findInfos)
	}

	graph.String()
}