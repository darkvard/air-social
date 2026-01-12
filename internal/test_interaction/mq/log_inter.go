package mq

import "air-social/pkg"

func logInfo(component, action, msg string, args ...any) {
	pkg.LogTemplate(pkg.InfoLog, component, action, msg, args...)
}

func logError(component, action, msg string, args ...any) {
	pkg.LogTemplate(pkg.ErrorLog, component, action, msg, args...)
}
