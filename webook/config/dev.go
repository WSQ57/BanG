//go:build !k8s

// 没有k8s编译标签
package config

var Config = config{
	DB: DBConfig{
		// 本地连接
		DSN: "root:root@tcp(:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
