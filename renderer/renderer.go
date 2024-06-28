package renderer

type IRendererImplementation interface {
	Header()
	Footer()
}

type Renderer struct {
}
