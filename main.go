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
	"github.com/z7zmey/php-parser/printer"
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

func findInFile(path string, search string) []walkers.FindInfo {
	f, err := os.Open(path)
	if err != nil {
		return []walkers.FindInfo{} // TODO error handle
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 1
	finded := []walkers.FindInfo{}

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), search) {
			finded = append(finded, walkers.FindInfo{
				FileName: path,
				Line: scanner.Text(),
				LineNumber: line,
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
func findInDir(path string, search string) []walkers.FindInfo {
	finded := []walkers.FindInfo{}
	findedInFile := []walkers.FindInfo{}
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".php") {
			findedInFile = findInFile(path, search)
			finded = append(finded, findedInFile...)
		}
		return nil
	})
	return finded
}

// инициализируем граф
var graph structure.FindInfoGraph

// итерация заполнения графа
func addFindedToGraph(method string, parent structure.Node, findInfoMap map[string][]walkers.FindInfo) {

	for fileName, findInfos := range findInfoMap {
		fmt.Println(fileName)

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

			// добавляем найденное в граф
			node := structure.Node{findInfo}
			graph.AddNode(&node)
			graph.AddEdge(&node, &parent)

			nodeFinder := walkers.NodeFinder{
				Writer:        os.Stdout,
				Positions:     parser.GetPositions(),
				FindSignature: method,
				FindInfo:      &findInfo,
			}

			rootNode := parser.GetRootNode()
			rootNode.Walk(nodeFinder)

			// вывод в буфер
			buf := new(bytes.Buffer)
			p := printer.NewPrinter(buf, "")
			p.Print(findInfo.Node)

			// TODO: надо разрезолвить неймспейс найденого метода, чтобы использовать его при дальнейших чеках
			fmt.Println(buf.String())
		}
	}
}

func main() {
	conf := loadConfig()
	phpShtormReference := "\\App\\Model\\PO\\PayItem::removeByOrderIdXXX"

	parts := strings.Split(phpShtormReference, "::")
	namespace, method := parts[0], parts[1]
	fmt.Println(namespace)

	finded := []walkers.FindInfo{}

	for _, dir := range conf.ScanDirs {
		var path = strings.Join([]string{conf.RootDir, dir.Name}, "/")
		fmt.Println("Dir:", path)
		finded = findInDir(path, strings.Join([]string{"->", method, "("}, ""))

		fmt.Println("Count finded:", len(finded))
	}

	// делаем мапу, группируя найденное по именам файлов
	FindInfoMap := make(map[string][]walkers.FindInfo)
	for _, findInfo := range finded {
		FindInfoMap[findInfo.FileName] = append(FindInfoMap[findInfo.FileName], findInfo)
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

	addFindedToGraph(method, rootNode, FindInfoMap)

	graph.String()
}