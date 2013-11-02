package gos

const (
	CACHE_NOT_FOUND int = 0
	CACHE_FOUND     int = 1
	CACHE_DISABLED  int = -1
)

var (
	RenderNothing = &EmptyRender{}

	B_DOT          = []byte(".")
	B_HTML_SUBFIX  = []byte(".html")
	B_SLASH        = []byte("/")
	B_EQUAL        = []byte("=")
	B_QUOTE        = []byte("\"")
	B_SINGLE_QUOTE = []byte("'")
	B_SPACE        = []byte(" ")

	B_HTML_BEGIN      = []byte("<!DOCTYPE HTML>\n<html>\n")
	B_HTML_END        = []byte("</html>\n")
	B_HTML_BODY_BEGIN = []byte("\n<body>\n")
	B_HTML_BODY_END   = []byte("\n</body>\n")
)
