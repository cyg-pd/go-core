package render

import "sync"

type ErrorRenderFunc func(err error) (Renderable, bool)

var mapper []ErrorRenderFunc
var mu sync.Mutex

func RegisterErrorRender(f ErrorRenderFunc) {
	mu.Lock()
	defer mu.Unlock()

	mapper = append(mapper, f)
}
