package worker

import "os"

func assetDirectory() string {
	if path := os.Getenv("ASSET_DIR"); path != "" {
		return path
	}
	return "public"
}
