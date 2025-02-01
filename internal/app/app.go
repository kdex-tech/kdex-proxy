package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"kdex.dev/proxy/internal/util"
)

const (
	DefaultAppsPath = "/etc/kdex/apps"
)

type Apps []App

type App struct {
	Alias          string   `json:"alias"`
	Address        string   `json:"address"`
	Element        string   `json:"element"`
	Path           string   `json:"path"`
	Targets        []Target `json:"targets"`
	RequiredScopes []string `json:"requiredScopes"`
}

type Target struct {
	Page      string `json:"page"`
	Container string `json:"containerId,omitempty"`
}

type AppManager struct {
	Apps Apps
}

func NewAppManagerFromEnv() *AppManager {
	appsPath := os.Getenv("APPS_PATH")
	if appsPath == "" {
		appsPath = DefaultAppsPath
		log.Printf("Defaulting module_dir to %s", appsPath)
	}

	appsFile, err := os.ReadFile(appsPath)
	if err != nil {
		log.Fatalf("Failed to read apps directory: %v", err)
	}

	var apps Apps
	if err = json.Unmarshal(appsFile, &apps); err != nil {
		log.Fatalf("Failed to unmarshal apps: %v", err)
	}

	for _, app := range apps {
		if err := ValidateApp(app); err != nil {
			log.Fatalf("Invalid app: %v", err)
		}
	}

	return &AppManager{
		Apps: apps,
	}
}

func (m *AppManager) GetApps() Apps {
	return m.Apps
}

func (m *AppManager) GetAppsForPage(page string) Apps {
	var apps Apps
	for _, app := range m.Apps {
		for _, appTarget := range app.Targets {
			if appTarget.Page == page {
				apps = append(apps, app)
			}
		}
	}
	return apps
}

func ValidateApp(app App) error {
	if app.Address == "" {
		return fmt.Errorf("App address is required")
	}
	if app.Element == "" {
		return fmt.Errorf("App element is required")
	}
	if app.Path == "" {
		return fmt.Errorf("App path is required")
	}
	if len(app.Targets) == 0 {
		return fmt.Errorf("App must have at least one target")
	}
	for _, target := range app.Targets {
		if target.Page == "" {
			return fmt.Errorf("App targets page is required")
		}
	}

	if app.Alias == "" {
		app.Alias = util.RandStringBytes(4)
	}

	return nil
}
