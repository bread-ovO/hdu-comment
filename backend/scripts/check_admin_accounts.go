package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hdu-dp/backend/internal/config"
	"github.com/hdu-dp/backend/internal/database"
	"github.com/hdu-dp/backend/internal/models"
)

func main() {
	fmt.Println("=== 管理员账户检查工具 ===")
	fmt.Println()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("⚠️  配置加载失败: %v\n", err)
		fmt.Println("使用默认配置继续检查...")
		cfg = &config.Config{}
		cfg.Admin.Email = "admin@example.com"
	}

	// 设置默认JWT密钥用于检查
	os.Setenv("APP_AUTH_JWT_SECRET", "check-admin-secret-key")

	// 初始化数据库
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}

	// 检查所有用户
	var allUsers []models.User
	result := db.Find(&allUsers)
	if result.Error != nil {
		log.Fatalf("❌ 查询用户失败: %v", result.Error)
	}

	fmt.Printf("📊 数据库中共有 %d 个用户\n", len(allUsers))
	fmt.Println()

	// 显示所有用户
	fmt.Println("=== 所有用户列表 ===")
	for i, user := range allUsers {
		fmt.Printf("%d. ID: %s\n", i+1, user.ID)
		fmt.Printf("   邮箱: %s\n", user.Email)
		fmt.Printf("   显示名称: %s\n", user.DisplayName)
		fmt.Printf("   角色: %s\n", user.Role)
		fmt.Printf("   创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

		if user.Role == "admin" {
			fmt.Printf("   ✅ 这是管理员账户\n")
		}
		fmt.Println()
	}

	// 检查管理员账户
	var adminUsers []models.User
	db.Where("role = ?", "admin").Find(&adminUsers)

	fmt.Println("=== 管理员账户统计 ===")
	if len(adminUsers) == 0 {
		fmt.Println("❌ 没有找到管理员账户")
		fmt.Println()
		fmt.Println("💡 解决方案:")
		fmt.Println("1. 确保环境变量 ADMIN_EMAIL 和 ADMIN_PASSWORD 已设置")
		fmt.Println("2. 重启应用以创建管理员账户")
		fmt.Println("3. 或者手动创建管理员账户")
	} else {
		fmt.Printf("✅ 找到 %d 个管理员账户:\n", len(adminUsers))
		for _, admin := range adminUsers {
			fmt.Printf("   📧 %s (%s)\n", admin.Email, admin.DisplayName)
		}
	}

	// 检查配置中的管理员邮箱
	fmt.Println()
	fmt.Println("=== 配置检查 ===")
	fmt.Printf("配置中的管理员邮箱: %s\n", cfg.Admin.Email)
	if cfg.Admin.Email == "" {
		fmt.Println("⚠️  环境变量 ADMIN_EMAIL 未设置")
		fmt.Println("建议设置: export APP_ADMIN_EMAIL=your-admin@email.com")
	}

	// 检查特定管理员用户
	if cfg.Admin.Email != "" {
		var specificAdmin models.User
		err := db.Where("email = ?", cfg.Admin.Email).First(&specificAdmin).Error
		if err != nil {
			fmt.Printf("❌ 配置中的管理员邮箱 %s 不存在\n", cfg.Admin.Email)
		} else {
			fmt.Printf("✅ 配置中的管理员邮箱 %s 存在，角色: %s\n",
				cfg.Admin.Email, specificAdmin.Role)
			if specificAdmin.Role != "admin" {
				fmt.Printf("⚠️  但该用户角色是 '%s' 而不是 'admin'\n", specificAdmin.Role)
			}
		}
	}

	fmt.Println()
	fmt.Println("=== 下一步操作 ===")
	if len(adminUsers) == 0 {
		fmt.Println("1. 设置环境变量:")
		fmt.Println("   export APP_ADMIN_EMAIL=admin@yourdomain.com")
		fmt.Println("   export APP_ADMIN_PASSWORD=secure-password")
		fmt.Println("2. 重启应用")
		fmt.Println("3. 重新登录获取管理员令牌")
	} else {
		fmt.Println("1. 使用现有管理员账户登录")
		fmt.Println("2. 确保JWT令牌包含正确的角色信息")
		fmt.Println("3. 测试管理员权限")
	}
}
