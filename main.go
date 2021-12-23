package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var counter = struct {
	sync.RWMutex
	mapDatabase map[string]int
	mapIp       map[string]int
	mapHost     map[string]int
	mapKey      map[string]int
	mapCommand  map[string]int
}{
	mapDatabase: make(map[string]int),
	mapIp:       make(map[string]int),
	mapHost:     make(map[string]int),
	mapKey:      make(map[string]int),
	mapCommand:  make(map[string]int),
}

var sum = 0
var count = 0

func main() {
	Monitor()
}

func Monitor() {
	//ScannerInput()
	//printNode()
	go printLoop()
	scanner()
	fmt.Println("exit")
}
func printLoop() {
	for {
		printNode()
		time.Sleep(time.Duration(5) * time.Second)

	}
}

func scanner() {
	scanner := bufio.NewReader(os.Stdin)
	for {
		line, err := scanner.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%#v\n", line)
				break
			}
			panic(err)
		}
		sum = sum + 1
		readerRedisMessage(line)
	}
}

func printNode() {
	count++
	counter.RLock()
	var sb strings.Builder
	sortData("map_database", counter.mapDatabase, &sb)
	sortData("map_command", counter.mapCommand, &sb)
	sortData("map_key", counter.mapKey, &sb)
	sortData("map_ip", counter.mapIp, &sb)
	sortData("map_host", counter.mapHost, &sb)
	var avg = 0
	if sum != 0 {
		avg = sum / (5 * count)
	}
	sb.WriteString("====(sum:" + strconv.Itoa(sum) + " / avg:" + strconv.Itoa(avg) + ")====\n")
	fmt.Print(sb.String())
	counter.RUnlock()
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func sortData(title string, barVal map[string]int, sb *strings.Builder) {

	sb.WriteString("====== [" + title + "] ======\n")
	pl := make(PairList, len(barVal))
	i := 0
	for k, v := range barVal {
		pl[i] = Pair{k, v}
		i++
	}
	//先排序
	sort.Sort(sort.Reverse(pl))
	var td = pl
	//取前十
	if len(pl) > 10 {
		td = pl[0:10]
	}
	for _, k := range td {
		var node = k.Key + "：" + strconv.Itoa(k.Value) + " avg :" + strconv.Itoa(k.Value/count/5)
		node = strings.Replace(node, "\n", "", -1)
		sb.WriteString(node + "\n")
	}

}

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func ScannerInput() {
	fi, err := os.Open("D:/proj/goredis/m2.log")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	i := 0
	for {
		i++
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		readerRedisMessage(string(a))
	}
}

func readerRedisMessage(message string) {
	counter.Lock()
	messageList := strings.Split(message, " ")
	if len(messageList) < 4 {
		counter.Unlock()
		return
	}
	for i := 0; i < len(messageList); i++ {
		element := messageList[i]
		element = strings.Trim(element, "[")
		element = strings.Trim(element, "]")
		element = strings.Trim(element, "\"")
		_, err := strconv.ParseFloat(messageList[0], 2)
		if err != nil || len(messageList[0]) != 17 {
			continue
		}
		if i == 1 {
			counter.mapDatabase[element] = counter.mapDatabase[element] + 1
		} else if i == 2 {
			if element == "lua" {
				continue
			}
			//ip端口
			host := strings.Split(element, ":")
			ip := host[0]
			counter.mapIp[ip] = counter.mapIp[ip] + 1
			counter.mapHost[element] = counter.mapHost[element] + 1

		} else if i == 3 {
			counter.mapCommand[element] = counter.mapCommand[element] + 1
		} else if i == 4 {
			counter.mapKey[element] = counter.mapKey[element] + 1
		}
	}
	counter.Unlock()
}
