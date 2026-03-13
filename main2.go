package main

import (
	"encoding/json"
	"fmt"
	"github.com/opdss/go-helper/path"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"rest_demo/internal/migration"
	"rest_demo/internal/service"
	"rest_demo/pkg/cache"
	"rest_demo/pkg/cfgstruct"
	"rest_demo/pkg/db"
	"rest_demo/pkg/jwt"
	"rest_demo/pkg/log"
	"rest_demo/pkg/payment/wechat"
	"rest_demo/pkg/process"
	"rest_demo/pkg/redis"
	"rest_demo/pkg/server/http"
	"rest_demo/wire"
)

type Config struct {
	Api       http.Config
	Log       log.Config
	Db        db.MsConfig
	JWT       jwt.Config
	Redis     redis.Config
	Service   service.Config
	WechatPay wechat.WeChatPayConfig
	Cache     cache.Config
}

var (
	cfg          Config
	migrationCfg struct {
		Admin migration.Config
		Db    db.MsConfig
		Log   log.Config
	}
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:   "main",
		Short: "admin serverf",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "运行",
		RunE:  cmdRun,
	}
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "查看当前所有配置",
		RunE:  cmdConfig,
	}
	setupCmd = &cobra.Command{
		Use:         "setup",
		Short:       "初始化配置",
		RunE:        cmdSetup,
		Annotations: map[string]string{"type": "setup"},
	}
	migrationCmd = &cobra.Command{
		Use:   "migration",
		Short: "同步数据表",
		RunE:  cmdMigration,
	}
	// migrationCmd = &cobra.Command{
	// 	Use:   "migration",
	// 	Short: "数据库迁移初始化等操作,可选[init_tables|init_admin|mock|default]",
	// 	Args:  cobra.ExactArgs(1),
	// 	RunE:  cmdMigration,
	// }
	// versionCmd = &cobra.Command{
	// 	Use:   "version",
	// 	Short: "查看版本信息",
	// 	RunE:  cmdVersion,
	// }
	// execCmd = &cobra.Command{
	// 	Use:   "exec",
	// 	Short: "执行自定义方法",
	// 	RunE:  cmdExec,
	// }
)

type noCopy struct{}

// Lock 和 Unlock 方法空实现，仅为了满足 sync.Locker 接口
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type MyResource struct {
	noCopy noCopy
}

// @title Hobi-admin API
// @version v1.0.0
// @BasePath /api
// @description   hobi-admin 后台管理接口文档
// @description   接口文档地址：http://localhost/api/swagger/index.html
// @description   测试接口地址：http://localhost/api/
// @description   生产接口地址：https://localhost/api/
// @description
// @description  约定规范
// @description   1. 接口统一采用RESTful风格
// @description   2. 接口请求参数和响应参数名统小写一下划线风格
// @description   3. 接口路径统一小写下划线风格
// @description   4. 授权认证采用基于jwt的Bearer认证方案
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main2() {
	defaultConfig := path.ApplicationDir("rest_demo", process.DefaultCfgFilename)
	cfgstruct.SetupFlag(rootCmd, &configFile, "config", defaultConfig, "配置文件")
	//根据环境读取默认配置
	defaults := cfgstruct.DefaultsFlag(rootCmd)
	// 当前程序所在目录
	currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	rootDir := cfgstruct.ConfigVar("ROOT", currentDir)
	// 设置系统的HOME变量
	envHome := cfgstruct.ConfigVar("HOME", os.Getenv("HOME"))
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(migrationCmd)
	// rootCmd.AddCommand(migrationCmd)
	// rootCmd.AddCommand(versionCmd)
	// rootCmd.AddCommand(execCmd)
	process.Bind(runCmd, &cfg, defaults, cfgstruct.ConfigFile(configFile), envHome, rootDir)
	process.Bind(configCmd, &cfg, defaults, cfgstruct.ConfigFile(configFile), envHome, rootDir)
	process.Bind(setupCmd, &cfg, defaults, cfgstruct.ConfigFile(configFile), envHome, cfgstruct.SetupMode(), rootDir)
	process.Bind(migrationCmd, &migrationCfg, defaults, cfgstruct.ConfigFile(configFile), envHome, rootDir)
	// process.Bind(migrationCmd, &migrationCfg, defaults, cfgstruct.ConfigFile(configFile), envHome, rootDir)
	// process.Bind(versionCmd, &struct{}{}, defaults)
	// process.Bind(execCmd, &cfg, defaults, cfgstruct.ConfigFile(configFile), envHome, rootDir)
	process.Exec(rootCmd)
}

// cmdConfig 查看系统配置
func cmdConfig(cmd *cobra.Command, args []string) error {
	fmt.Printf("当前运行环境：[%s]\n", cfgstruct.DefaultsType())
	fmt.Println("当前配置文件路径：", configFile)
	output, _ := json.MarshalIndent(cfg, "", " ")
	fmt.Println(string(output))
	return nil
}

// cmdRun 运行
func cmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx, _ := process.Ctx(cmd)

	//notice.Init(&cfg.Notice)
	_log := log.NewLog(&cfg.Log)
	app, fn, err := wire.NewApp(ctx, _log, &cfg.Db, &cfg.Api, &cfg.Service, &cfg.JWT, &cfg.Redis, &cfg.WechatPay, &cfg.Cache)
	if err != nil {
		return err
	}
	defer fn()

	if err = app.Run(ctx); err != nil {
		_log.Error("app.Run error", zap.Error(err))
	}

	return nil
}

// cmdSetup 初始化数据库
func cmdSetup(cmd *cobra.Command, args []string) error {
	return process.SaveConfig(cmd, configFile)
}

// cmdMigration 数据库迁移初始化
func cmdMigration(cmd *cobra.Command, args []string) error {
	_log := log.NewLog(&migrationCfg.Log)
	db, err := db.NewDB(migrationCfg.Db.Master)
	if err != nil {
		return err
	}
	fmt.Println("运行数据库[", migrationCfg.Db.Master.Driver, "]：", migrationCfg.Db.Master.Dsn)
	mr := migration.NewMigration(&migrationCfg.Admin, _log, db)
	return mr.InitTables()
}
