package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

func writeCache(filename string, data map[string]interface{}, expireTimestamp int64) error {
	// 将数据存储为 JSON 格式到本地文件
	data["expire_at"] = expireTimestamp
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建缓存文件失败: %v", err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return fmt.Errorf("将数据编码为 JSON 失败: %v", err)
	}

	return nil
}

func getCached(filename string) (map[string]interface{}, error) {
	// 从本地缓存文件中读取数据
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 缓存文件不存在
		}
		return nil, fmt.Errorf("打开缓存文件失败: %v", err)
	}
	defer file.Close()

	var cache map[string]interface{}
	err = json.NewDecoder(file).Decode(&cache)
	if err != nil {
		return nil, fmt.Errorf("解析 JSON 缓存数据失败: %v", err)
	}

	expireTimestamp, ok := cache["expire_at"].(float64)
	if !ok {
		return nil, errors.New("缓存数据中的 expire_at 值无效")
	}
	if int64(expireTimestamp) <= time.Now().Unix() {
		log.Printf("缓存已过期，删除文件: %s\n", filename)
		return nil, nil // 缓存已过期
	}

	return cache, nil
}
