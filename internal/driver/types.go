package driver

import (
	"context"
	"time"
	
	"github.com/kazuph/mcp-android-chrome/internal/loader"
)


// DriverConfig holds common configuration for all drivers
type DriverConfig struct {
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
	Debug   bool          `json:"debug"`
	File    string        `json:"file"`
}

// Driver interface defines the common functionality for all drivers
type Driver interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetURL() string
	CheckEnvironment() error
	LoadTabs(ctx context.Context) ([]loader.Tab, error)
}

// AndroidConfig extends DriverConfig with Android-specific options
type AndroidConfig struct {
	DriverConfig
	Socket      string        `json:"socket"`
	Wait        time.Duration `json:"wait"`
	SkipCleanup bool          `json:"skipCleanup"`
}

// IOSConfig extends DriverConfig with iOS-specific options  
type IOSConfig struct {
	DriverConfig
	Wait time.Duration `json:"wait"`
}

// RestoreDriver interface for tab restoration functionality
type RestoreDriver interface {
	Driver
	RestoreTabs(ctx context.Context, tabs []loader.Tab) error
}