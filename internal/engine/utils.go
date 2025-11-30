package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"
)

// LoadRootConfig carrega um arquivo JSON simples de config.
func LoadRootConfig(path string) (*kbx.RootConfig, error) {
	if path == "" {
		path = os.ExpandEnv(kbx.DefaultCanalizeDSConfigPath)
	}

	data, err := os.ReadFile(os.ExpandEnv(path))
	if err != nil {
		return nil, err
	}
	// var cfg *RootConfig
	cfgMp := types.NewMapper(&kbx.RootConfig{}, path)
	cfgObj, err := cfgMp.Deserialize(data, filepath.Ext(path)[1:])
	if err != nil {
		return nil, err
	}
	if cfgObj != nil {
		return cfgObj, nil
	}

	newPath := filepath.Join(os.ExpandEnv(kbx.DefaultConfigDir), "canalize_ds", "config", filepath.Base(path))
	cfgMpC := types.NewMapperType(cfgMp.GetObject(), os.ExpandEnv(newPath))
	cfgMpC.SerializeToFile(filepath.Ext(path)[1:])
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config file not found at %s", path)
	}

	return nil, errors.New("failed to deserialize root config")
}

// SaveRootConfig salva o arquivo JSON.
func SaveRootConfig(cfg *kbx.RootConfig) error {
	if cfg.FilePath == "" {
		return errors.New("root config FilePath is empty")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.FilePath, data, 0o640)
}

// GetDefaultConfigPath calcula o path padrão $HOME/.canalize/database/canalize_db/config.json
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(
		home,
		".canalize",
		"database",
		"canalize_db",
		"config.json",
	), nil
}

// GenerateRandomPassword é só um helper simples (pode trocar pela sua versão oficial).
func GenerateRandomPassword(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-" // pragma: allowlist secret
	buf := make([]byte, n)
	f, err := os.Open("/dev/urandom")
	if err != nil {
		// fallback tosco, mas ok
		for i := range buf {
			buf[i] = alphabet[i%len(alphabet)]
		}
		return string(buf)
	}
	defer f.Close()
	_, _ = f.Read(buf)
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf)
}

// GenerateDefaultPostgresConfig gera uma única config de Postgres básica.
func GenerateDefaultPostgresConfig() *kbx.RootConfig {
	pass := GenerateRandomPassword(40)

	db := &kbx.DBConfig{
		// ID:        "canalize_db",
		// Name:      "canalize_db",
		// IsDefault: true,
		Enabled: kbx.BoolPtr(true),
		// Type:      DBTypePostgres,
		Host:   "127.0.0.1",
		Port:   "5432",
		User:   "canalize_adm",
		Pass:   pass,
		DBName: "canalize_db",
		Schema: "public",
		Options: map[string]any{
			"sslmode":           "disable",
			"max_connections":   50,
			"connect_timeout":   10,
			"application_name":  "canalize_ds",
			"pool_max_lifetime": "30m",
		},
	}

	// DSN simples, o driver pode refinar
	db.DSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s", // pragma: allowlist secret
		db.User,
		db.Pass,
		db.Host,
		db.Port,
		db.DBName,
		db.Options["sslmode"],
	)

	return &kbx.RootConfig{
		Name:      "canalize_ds",
		Enabled:   kbx.BoolPtr(true),
		Databases: []*kbx.DBConfig{db},
	}
}

// BootstrapDatabaseManager é o entrypoint que o main do DS pode chamar.
func BootstrapDatabaseManager(ctx context.Context, logger *logz.LoggerZ, cfgPath string) (*kbx.RootConfig, error) {
	mgr := NewDatabaseManager(logger)

	root, err := mgr.LoadOrBootstrap(cfgPath)
	if err != nil {
		return nil, err
	}

	if err := mgr.InitFromRootConfig(ctx, root); err != nil {
		return nil, err
	}

	return root, nil
}
