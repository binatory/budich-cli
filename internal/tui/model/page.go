package model

// FIXME call it Screen instead ?
type PageEnum string

const (
	PageList   PageEnum = "PageList"
	PageSearch PageEnum = "PageSearch"
)

func (pe PageEnum) String() string {
	return string(pe)
}
