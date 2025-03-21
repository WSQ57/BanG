//go:build k8s

// 使用k8s编译标签
package config

var Config = config{
	DB: DBConfig{
		// 本地连接
		DSN: "root:root@tcp(webook-mysql:3309)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:11479",
	},
}
