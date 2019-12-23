package hotypes

import (
	"encoding/xml"
	//	dat "github.com/joergreinhardt/gatwd/data"
	//	fnc "github.com/joergreinhardt/gatwd/functions"
)

// needs to be a recursively defined struct in order to implement xml.decoder,
// since it needs struct tags to map fields to values
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []XMLNode  `xml:",any"`
}

func (n *XMLNode) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node XMLNode

	return dec.DecodeElement((*node)(n), &start)
}

func walkXML(nodes []XMLNode, f func(XMLNode) bool) {
	for _, n := range nodes {
		if f(n) {
			walkXML(n.Nodes, f)
		}
	}
}
