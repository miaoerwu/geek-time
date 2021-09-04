package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	gorma "github.com/hhxsv5/go-redis-memory-analysis"
)

var client redis.UniversalClient
var ctx context.Context

const (
	ip   string = "127.0.0.1"
	port uint16 = 6379
)

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%v:%v", ip, port),
		Password:     "",
		DB:           0,
		PoolSize:     128,
		MinIdleConns: 100,
		MaxRetries:   3,
	})

	ctx = context.Background()
}

func main() {
	write(10000, "10k_len10", generateValue(10))
	write(50000, "50k_len10", generateValue(10))
	write(500000, "500k_len10", generateValue(10))

	write(10000, "10k_len1000", generateValue(1000))
	write(50000, "50k_len1000", generateValue(1000))
	write(500000, "500k_len1000", generateValue(1000))

	write(10000, "10k_len5000", generateValue(5000))
	write(50000, "50k_len5000", generateValue(5000))
	write(500000, "500k_len5000", generateValue(5000))

	analysis()
}

func write(num int, key, value string) {
	for i := 0; i < num; i++ {
		k := fmt.Sprintf("%s:%v", key, i)
		cmd := client.Set(ctx, k, value, redis.KeepTTL)
		err := cmd.Err()
		if err != nil {
			fmt.Println(cmd.String())
		}
	}
}

func generateValue(size int) string {
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		arr[i] = 'a'
	}
	return string(arr)
}

func analysis() {
	analysis, err := gorma.NewAnalysisConnection(ip, port, "")
	if err != nil {
		fmt.Println("something wrong:", err)
		return
	}
	defer analysis.Close()

	analysis.Start([]string{":"})

	err = analysis.SaveReports("./analysis_reports")
	if err == nil {
		fmt.Println("complete")
	} else {
		fmt.Println("error:", err)
	}
}
