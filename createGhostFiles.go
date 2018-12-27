package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type line struct {
	ting []int
	fan  string
}

var repeated map[string][]*line
var fileName = 0

func main() {

	for i := 10; i < 19; i++ {
		fileName = i
		repeated = make(map[string][]*line)
		b, e := ioutil.ReadFile("D:/HuFiles/" + strconv.Itoa(fileName) + ".txt")
		if e != nil {
			fmt.Println("read file error")
			return
		}
		// 读取所有行
		array := strings.Split(string(b), "\n")
		for _, value := range array {
			if value == "" {
				continue
			}
			index := strings.Index(value, "*")
			if index == -1 {
				continue
			}

			head := make([]int, index)
			for i := 0; i < index; i++ {
				tmpNum, _ := strconv.Atoi(string(value[0:index][i]))
				head[i] = tmpNum
			}
			left := ""
			if index != -1 {
				left = value[index+1:]
			}
			reduceNum(head, left)
		}

		writeFileBymap()
		time.Sleep(1000)
	}

}

func writeFileBymap() {
	file1, _ := os.OpenFile("D:/HuFiles/"+strconv.Itoa(fileName+1)+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	for key, lines := range repeated {
		finalStr := key + "*"

		for _, oneLine := range lines {
			finalStr += "|" + oneLine.fan + "|"
			for index, ti := range oneLine.ting {
				if index == len(oneLine.ting)-1 {
					finalStr += strconv.Itoa(ti) + " "
				} else {
					finalStr += strconv.Itoa(ti) + ","
				}
			}
		}

		file1.WriteString(finalStr + "\n")
	}

	file1.Close()
}
func reduceNum(srcList []int, left string) {
	for j := 0; j < len(srcList); j++ {
		if srcList[j] == 0 {
			continue
		}
		tmpList := make([]int, len(srcList))
		copy(tmpList, srcList)
		if srcList[j] >= 1 {
			tmpList[j] -= 1
		}
		addMap(tmpList, left, j+1)
	}
}

//merge
func addMap(srcList []int, left string, addNum int) {
	// 唯一主键 索引

	addLine := make([]*line, 0)

	fanList := strings.Split(left, " ")

	for _, oneFan := range fanList {
		index1 := strings.Index(oneFan, "|") + 1
		index2 := strings.LastIndex(oneFan, "|") + 1
		if 0 == index1 {
			continue
		}
		fanName := oneFan[index1 : index2-1]

		ting := make([]int, 0)
		flag := false

		// 添加原来的
		for _, num := range strings.Split(oneFan[index2:], ",") {
			intNum, err := strconv.Atoi(num)
			if err != nil {
				continue
			}
			ting = append(ting, intNum)
			if intNum == addNum {
				flag = true
			}
		}

		//添加新增的
		if !flag {
			ting = append(ting, addNum)
		}

		fanFlag := false
		for _, oneLine := range addLine {
			if oneLine.fan == fanName {
				fanFlag = true
			}
		}
		if !fanFlag {
			addLine = append(addLine, &line{
				ting: ting,
				fan:  fanName})
		}

	}

	curline := ""
	for _, value := range srcList {
		curline += strconv.Itoa(value)
	}

	if _, ok := repeated[curline]; !ok {
		repeated[curline] = addLine
		return
	}

	// 添加到map

	for _, addLine := range addLine {
		realNot := false
		for index, oneLine := range repeated[curline] {
			if oneLine.fan == addLine.fan {
				for _, t2 := range addLine.ting {
					curFlag := false
					for _, t1 := range oneLine.ting {
						if t1 == t2 {
							curFlag = true
							break
						}
					}
					if curFlag == false {
						repeated[curline][index].ting = append(repeated[curline][index].ting, t2)
					}
				}
				realNot = true
			}
		}
		if realNot == false {
			repeated[curline] = append(repeated[curline], addLine)
		}
	}

	//findindex := -1
	//for index, value := range repeated[curline] {
	//	if strings.Split(value.fan, "*")[0] == strings.Split(left, "*")[0] {
	//		findindex = index
	//	}
	//}
	//if findindex != -1 {
	//	flag := false
	//	for _, tvalue := range repeated[curline][findindex].ting {
	//		if tvalue == addNum {
	//			flag = true
	//		}
	//	}
	//	if flag {
	//		return
	//	}
	//	if flag == false {
	//		repeated[curline][findindex].ting = append(repeated[curline][findindex].ting, addNum)
	//		return
	//	}
	//}
	//
	//repeated[curline] = append(repeated[curline], &line{
	//	ting: []int{addNum},
	//	fan:  left,
	//	//head: head,
	//})

	return
}
