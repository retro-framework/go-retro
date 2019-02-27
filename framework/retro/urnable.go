package retro

// URNAble is an interface for references, strong or weak
// where things carry a name. A concrete implemnentation of
//
type URNAble interface {
	URN() URN
}
