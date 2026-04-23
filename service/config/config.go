package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var AppConfig *Config

type Config struct {
	MySQL  MySQLConfig  `yaml:"mysql"`
	JWT    JWTConfig    `yaml:"jwt"`
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	// Port HTTP 监听端口，对应 Gin `Run(":端口")`。
	Port int `yaml:"port"`
}

type MySQLConfig struct {
	Prefix   string `yaml:"prefix"`
	Port     string `yaml:"port"`
	Config   string `yaml:"config"`
	DBName   string `yaml:"dbname"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Path     string `yaml:"path"`
}

type JWTConfig struct {
	Secret               string `yaml:"secret"`
	Issuer               string `yaml:"issuer"`
	ExpireHours          int    `yaml:"expire_hours"`
	RefreshBeforeMinutes int    `yaml:"refresh_before_minutes"`
}

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	searchDirs := collectConfigDirs()
	for _, dir := range searchDirs {
		viper.AddConfigPath(dir)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf(
			"读取配置文件失败: %v\n"+
				"请任选其一：\n"+
				"1) 将 config.yaml 放在可执行文件同级的 config/ 目录下；\n"+
				"2) 设置环境变量 CONFIG_PATH 为包含 config.yaml 的目录；\n"+
				"3) 在进程守护里把工作目录设为 service 目录。",
			err,
		)
	}

	AppConfig = &Config{}

	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct:%v", err)
	}
}

func collectConfigDirs() []string {
	var dirs []string
	seen := map[string]bool{}

	add := func(p string) {
		if p == "" {
			return
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return
		}
		if seen[abs] {
			return
		}
		seen[abs] = true
		dirs = append(dirs, abs)
	}

	if env := os.Getenv("CONFIG_PATH"); env != "" {
		add(env)
	}

	if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, "config"))
		add(wd)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		add(filepath.Join(exeDir, "config"))
		add(filepath.Join(exeDir, "..", "config"))
		add(exeDir)
	}

	add("./config")
	add(".")

	return dirs
}
