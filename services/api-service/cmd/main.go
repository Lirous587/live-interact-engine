package main

import (
	_ "live-interact-engine/shared/logger"

	"go.uber.org/zap"
)

func main() {
	zap.L().Info("test zap")
}
