package util

import (
	"bytes"
	"errors"
	"golang.org/x/net/html"
	"image/color"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".gifv"}
)

func FileExtMatches(s []string, file string) bool {
	found := false
	file = strings.ToLower(file)

	for _, e := range s {
		if filepath.Ext(file) == e {
			found = true
			break
		}
	}

	return found
}

type extractNodeCondition func(string) bool

// ExtractNode will select the first node to match extractNodeCondition, for example
// res, err := ExtractNode(string(content), func(str string) bool { return str == "title" })
func ExtractNode(content string, fn extractNodeCondition) (*html.Node, error) {
	doc, _ := html.Parse(strings.NewReader(string(content)))
	var n *html.Node
	var crawler func(*html.Node)

	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && fn(node.Data) {
			n = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if n != nil {
		return n, nil
	}
	return nil, errors.New("missing matching tag in the node tree")
}

func ExtractNodeText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ExtractNodeText(c, buf)
	}
}

// GetUserMention will return a formatted user mention from an id
func GetUserMention(id int64) string {
	return "<@!" + strconv.FormatInt(id, 10) + ">"
}

// ConvertColorToInt32 will convert 3 uint8s into one int32
func ConvertColorToInt32(c color.RGBA) int32 {
	return int32((uint32(c.R) << 16) | (uint32(c.G) << 8) | (uint32(c.B) << 0))
}

// ParseHexColorFast will take a hex string, and convert it to a color.RGBA
func ParseHexColorFast(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, GenericError("ParseHexColorFast", "parsing \""+s+"\"", "missing #")
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = SyntaxError("ParseHexColorFast", s)
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = SyntaxError("ParseHexColorFast", s)
	}
	return
}