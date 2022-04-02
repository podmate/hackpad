// +build js

package process

import (
	"syscall/js"

	"github.com/hack-pad/hackpad/internal/interop"
)

var (
	jsGo = js.Global().Get("Go")
)

func (p *Process) JSValue() js.Value {
	return js.ValueOf(map[string]interface{}{
		"pid":   p.pid,
		"ppid":  p.parentPID,
		"error": interop.WrapAsJSError(p.err, "spawn"),
	})
}

func (p *Process) StartCPUProfile() error {
	return interop.StartCPUProfile(p.ctx)
}
