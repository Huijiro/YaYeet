package files

import "path/filepath"

const GameFolder = "Games/Voices of the Void"

func GameFolderPath(homeDir string) string {
	return filepath.Join(homeDir, GameFolder)
}
