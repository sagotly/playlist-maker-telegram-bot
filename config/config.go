package config

import (
	"io/ioutil"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Config — единая структура для всех микросервисов.
type Config struct {
	TgBotToken string `yaml:"tg_bot_token"`
}

// interpolateEnv выполняет замену шаблонов ${VAR_NAME} на значения переменных окружения.
func interpolateEnv(input string) string {
	re := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		groups := re.FindStringSubmatch(match)
		if len(groups) == 2 {
			if val, ok := os.LookupEnv(groups[1]); ok {
				return val
			}
		}
		// Если переменная не найдена, оставляем исходное значение
		return match
	})
}

// interpolateConfig выполняет интерполяцию переменных окружения во всём YAML-содержимом.
func interpolateConfig(data []byte) []byte {
	return []byte(interpolateEnv(string(data)))
}

// LoadConfig загружает конфигурацию из YAML-файла с учётом подстановки переменных окружения.
func LoadConfig(configPath string) (*Config, error) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	interpolated := interpolateConfig(raw)

	var cfg Config
	if err := yaml.Unmarshal(interpolated, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
