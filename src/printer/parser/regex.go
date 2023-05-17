package parser

import "regexp"

var (
	G0_G1_G92    = regexp.MustCompile(`^G([01]|92)(\s|$)`)
	G28          = regexp.MustCompile(`^G28(\s|$)`)
	G90          = regexp.MustCompile(`^G90(\s|$)`)
	G91          = regexp.MustCompile(`^G91(\s|$)`)
	G92          = regexp.MustCompile(`^G92(\s|$)`)
	M106         = regexp.MustCompile(`^M106(\s|$)`)
	M107         = regexp.MustCompile(`^M107(\s|$)`)
	M112         = regexp.MustCompile(`^M112(\s|$)`)
	M118         = regexp.MustCompile(`^M118(\s|$)`)
	M18_M84_M410 = regexp.MustCompile(`^M(18|84|410)(\s|$)`)
	M220         = regexp.MustCompile(`^M220(\s|$)`)
	M220_M221    = regexp.MustCompile(`^M22[01](\s|$)`)
	M82          = regexp.MustCompile(`^M82(\s|$)`)
	M83          = regexp.MustCompile(`^M83(\s|$)`)
)
