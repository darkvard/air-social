package middleware

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"

	"air-social/pkg"
)

func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		err, ok := recovered.(error)
		if !ok {
			err = fmt.Errorf("%v", recovered)
		}

		// Capture the call stack. Skip the first 3 frames (runtime.Callers, this closure, middleware)
		panicSource := "unknown_source"
		pcs := make([]uintptr, 20)
		numFramesCaptured := runtime.Callers(3, pcs)
		frames := runtime.CallersFrames(pcs[:numFramesCaptured])

		for {
			frame, more := frames.Next()
			isRuntimeCode := strings.Contains(frame.File, "runtime/")
			isGinFrameworkCode := strings.Contains(frame.File, "gin-gonic/gin")

			if !isRuntimeCode && !isGinFrameworkCode {
				panicSource = fmt.Sprintf("%s:%d", frame.File, frame.Line)
				break
			}
			if !more {
				break
			}
		}

		pkg.Log().Errorw("panic recovered",
			"error", err,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"source", panicSource,
		)

		pkg.InternalError(c, "internal server error")
	})
}
