//go:build k8s

// 使用 k8s 这个编译标签

package config

var Config = config{
	DB: DBConfig{
		// 后面是mysql是因为默认没有webook这个数据库 如果想要的话 需要自己先创建
		DSN: "root:root@tcp(webook-mysql:11309)/mysql",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:11379",
	},
}
