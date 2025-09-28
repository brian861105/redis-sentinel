package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func getRedisServerInfo() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"host": "172.19.0.2",
			"port": 6379,
			"role": "master",
		},
		{
			"host": "172.19.0.3",
			"port": 6379,
			"role": "slave1",
		},
		{
			"host": "172.19.0.7",
			"port": 6379,
			"role": "slave2",
		},
	}
}

func main() {
	ctx := context.Background()

	redisInfos := getRedisServerInfo()

	for _, info := range redisInfos {
		host := info["host"].(string)
		port := info["port"].(int)
		role := info["role"].(string)

		rdb := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", host, port),
		})

		err := rdb.Ping(ctx).Err()
		if err != nil {
			fmt.Printf("Failed to connect to %s (%s:%d): %v\n", role, host, port, err)
		} else {
			fmt.Printf("Successfully connected to %s (%s:%d)\n", role, host, port)
		}

		result, err := rdb.Do(ctx, "ROLE").Result()
		if err != nil {
			fmt.Printf("Failed to execute ROLE command on %s (%s:%d): %v\n", role, host, port, err)
			rdb.Close()
			continue
		}
		fmt.Println(result)

		// ROLE 指令回傳陣列，第一個元素是角色
		if roleArray, ok := result.([]interface{}); ok && len(roleArray) > 0 {
			if role, ok := roleArray[0].(string); ok {
				fmt.Println(role)
			}
		}

		rdb.Close()
	}
}
