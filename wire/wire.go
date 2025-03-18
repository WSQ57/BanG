//go:build wireinject

// 让wire来注入代码
package wire

import (
	"dream/wire/repository"
	"dream/wire/repository/dao"

	"github.com/google/wire"
)

// InitRepository 使用 Wire 自动生成依赖注入代码
func InitRepository() *repository.UserRepository {
	wire.Build(
		InitDB,                       // 初始化数据库连接å
		dao.NewUserDAO,               // 创建 UserDAO 实例
		repository.NewUserRepository, // 创建 UserRepository 实例
	)
	return nil // Wire 会自动生成实际的返回值
}
