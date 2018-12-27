package config

import (
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"strings"
	"sync"
)

var (
	once     sync.Once
	instance *redis.Client
)

func main() {
	files, err := ioutil.ReadDir("D:/HuFiles/")
	if err != nil {
		fmt.Printf("read file err")
	}
	for _, file := range files {
		allData, err := ioutil.ReadFile("D:/HuFiles/" + file.Name())
		if err != nil {
			fmt.Printf("read file err")
			continue
		}
		allLine := strings.Split(string(allData), "\n")

		for _, line := range allLine {
			str := strings.Split(line, "*")
			if len(str) < 2 {
				fmt.Printf("split line err")
				continue
			}
			key := str[0]
			content := str[1]
			redisClient := GetClient()
			fileName := strings.Split(file.Name(), ".")[0]
			//realKey := fmt.Sprintf("table:%s:%d", fileName, key)
			redisClient.HSet(fileName, key, content)
			//a := redisClient.HGet(fileName, key)
			//fmt.Printf("split line err", a, b)
		}
	}
}

// 获取RedisClient
func GetClient() *redis.Client {
	addr := "127.0.0.1" + ":" + "6379"
	RedisPass := ""

	once.Do(func() {
		instance = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: RedisPass,
			DB:       1,
		})
	})

	return instance
}
