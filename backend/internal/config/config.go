package config

import (
	"fmt"
	"log"
	"sync"
	// 新增：导入strings包用于ENV键替换
	"strings"

	"github.com/spf13/viper"
)

// AppConfig 全局总配置
type AppConfig struct {
	App      AppSection          `mapstructure:"app"`
	DB       DBSection           `mapstructure:"db"`
	RabbitMQ RabbitMQSection     `mapstructure:"mq"`
	Redis    RedisSection        `mapstructure:"redis"`
	Chains   map[string]ChainCfg `mapstructure:"chains"` // 多链核心：key=链标识(eth_sepolia/bsc_test)
}

// AppSection/DBSection/RabbitMQSection/RedisSection 保持不变
type AppSection struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	HTTPPort int    `mapstructure:"httpPort"`
}
type DBSection struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	Timezone string `mapstructure:"timezone"`
}
type RabbitMQSection struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Vhost    string `mapstructure:"vhost"`
}
type RedisSection struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

// ChainCfg 单链配置（链基础配置+该链专属合约）
type ChainCfg struct {
	Enable    bool          `mapstructure:"enable"`    // 是否启用
	Basic     ChainBasicCfg `mapstructure:"basic"`     // 链基础配置
	Contracts []ContractCfg `mapstructure:"contracts"` // 该链下的所有合约
}

// ChainBasicCfg 链基础配置（每链独立，隔离多链）
type ChainBasicCfg struct {
	ChainID     int64  `mapstructure:"chain_id"`     // 链ID（int64防溢出）
	RPCUrl      string `mapstructure:"rpc_url"`      // 该链RPC地址
	WSUrl       string `mapstructure:"ws_url"`       // 该链WS地址（事件监听）
	Timeout     int    `mapstructure:"timeout"`      // 节点请求超时(秒)
	StartBlock  int    `mapstructure:"start_block"`  // 该链区块监听起始高度
	CorrectCron string `mapstructure:"correct_cron"` // 定时校正表达式
}

// ContractCfg 合约配置（归属具体链，与链一一绑定）
type ContractCfg struct {
	Name    string `mapstructure:"name"`    // 合约名称
	Address string `mapstructure:"address"` // 合约地址
	AbiPath string `mapstructure:"abiPath"` // 该链合约ABI路径（建议按链分目录）
	Handler string `mapstructure:"handler"` // 该链合约事件处理器（如eth_nft/bsc_nft）
}

var (
	instance *AppConfig
	once     sync.Once
	initErr  error
)

// Init 初始化配置（仅新增ENV前缀和键替换配置）
func Init() error {
	once.Do(func() {
		var cfg AppConfig
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		// 新增：设置ENV前缀+键替换规则（核心修改）
		viper.SetEnvPrefix("NFT")                              // ENV变量前缀，避免冲突
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 将YAML的.替换为ENV的_
		// 原有自动读取ENV保留
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			initErr = fmt.Errorf("读取配置失败: %w", err)
			return
		}
		if err := viper.Unmarshal(&cfg); err != nil {
			initErr = fmt.Errorf("解析配置失败: %w", err)
			return
		}

		// 多链初始化日志（打印所有链+合约信息）
		log.Printf("✅ 配置初始化成功 | 总链数: %d", len(cfg.Chains))
		for chainKey, chain := range cfg.Chains {
			log.Printf("⛓️  链[%s] | 链ID: %d | RPC: %s | 合约数: %d",
				chainKey, chain.Basic.ChainID, chain.Basic.RPCUrl, len(chain.Contracts))
			for _, c := range chain.Contracts {
				log.Printf("  📜 合约[%s] | 地址: %s | 处理器: %s", c.Name, c.Address, c.Handler)
			}
		}

		instance = &cfg
	})
	return initErr
}

// Get 获取全局配置实例
func Get() (*AppConfig, error) {
	if initErr != nil {
		return nil, initErr
	}
	if instance == nil {
		return nil, fmt.Errorf("配置未初始化，请先调用config.Init()")
	}
	return instance, nil
}

// DSN 生成PG数据库连接串
func (c *DBSection) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode, c.Timezone)
}

// BuildURL 生成RabbitMQ连接URL
func (c *RabbitMQSection) BuildURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s", c.User, c.Password, c.Host, c.Port, c.Vhost)
}
