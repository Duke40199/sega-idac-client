package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config is the read-only view of application configuration consumed by adapters.
type Config interface {
	GetGRPCCfg() GRPCConfig
	GetSegaClientCfg() SegaClientConfig
	GetAllNetClientCfg() AllNetClientConfig
}

func (c *AppConfig) GetGRPCCfg() GRPCConfig {
	return c.GRPCCfg
}
func (c *AppConfig) GetSegaClientCfg() SegaClientConfig {
	return c.SegaClientCfg
}
func (c *AppConfig) GetAllNetClientCfg() AllNetClientConfig {
	return c.AllNetClientConfig
}

// AppConfig holds all configuration
type AppConfig struct {
	GRPCCfg            GRPCConfig
	SegaClientCfg      SegaClientConfig
	AllNetClientConfig AllNetClientConfig
}

type GRPCConfig struct {
	Host string
	Port int
}

// Addr returns the host:port the gRPC server should listen on.
func (c GRPCConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type SegaClientConfig struct {
	SegaIDACHost string
	// sega cfgs
	GetListConstConfigURLPath string
	GetCurrentRoundUrlPath    string
	// time trail
	GetTimeTrailURLPath string
	// ob
	GetListOBRankingURLPath string
	// rankings
	GetTeamRankingUrlPath     string
	GetListPlayerGradeUrlPath string
}

type AllNetClientConfig struct {
	AllNetHost                  string
	GetListStoreLocationURLPath string
	// consts
	IDACGameCode        string
	EnglishLanguageCode string
}

// env getter funcs
func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*AppConfig, error) {
	_ = godotenv.Load()
	cfg := &AppConfig{
		GRPCCfg: GRPCConfig{
			Host: getEnv("GRPC_HOST", "0.0.0.0"),
			Port: getEnvAsInt("GRPC_PORT", 50051),
		},
		SegaClientCfg: SegaClientConfig{
			SegaIDACHost:              getEnv("SEGA_IDAC_HOST", ""),
			GetListConstConfigURLPath: getEnv("SEGA_IDAC_GET_LIST_CONST_CFG_URL_PATH", ""),
			GetTimeTrailURLPath:       getEnv("SEGA_IDAC_GET_TIME_TRAIL_URL_PATH", ""),
			GetListOBRankingURLPath:   getEnv("SEGA_IDAC_GET_LIST_OB_RANKING_URL_PATH", ""),
			GetCurrentRoundUrlPath:    getEnv("SEGA_IDAC_GET_CURRENT_ROUND_URL_PATH", ""),
			GetTeamRankingUrlPath:     getEnv("SEGA_IDAC_GET_TEAM_RANKING_URL_PATH", ""),
			GetListPlayerGradeUrlPath: getEnv("SEGA_IDAC_GET_LIST_PLAYER_GRADE_URL_PATH", ""),
		},
		AllNetClientConfig: AllNetClientConfig{
			AllNetHost:                  getEnv("ALLNET_HOST", ""),
			GetListStoreLocationURLPath: getEnv("ALLNET_GET_LIST_STORE_LOCATION_URL_PATH", ""),
			IDACGameCode:                getEnv("ALLNET_IDAC_GAME_CODE", ""),
			EnglishLanguageCode:         getEnv("ALLNET_EN_LANGUAGE_CODE", ""),
		},
	}

	return cfg, nil
}
