package httpgenerator

type HttpGenerator struct {
	*XContext
}

type TypesStructSpec struct {
	Name   string
	Fields []*FieldSpec
}

type FieldSpec struct {
	Name string
	Type string
	Tag  string
}

type ServiceSpec struct {
	Name    string
	Group   string
	Summary string
	Routes  []*RouteSpec
}

type RouteSpec struct {
	Method       string
	Path         string
	RustFulKey   string
	Name         string
	RequestType  string
	ResponseType string
	Summary      string
}

func NewGenerator(ctx *XContext) *HttpGenerator {
	return &HttpGenerator{
		ctx,
	}
}
