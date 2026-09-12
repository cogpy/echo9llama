package tray

import (
	"github.com/EchoCog/echollama/app/tray/commontray"
	"github.com/EchoCog/echollama/app/tray/wintray"
)

func InitPlatformTray(icon, updateIcon []byte) (commontray.OllamaTray, error) {
	return wintray.InitTray(icon, updateIcon)
}
