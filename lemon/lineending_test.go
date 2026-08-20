package lemon

import "testing"

func TestConvertLineEnding(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		option string
		want   string
	}{
		{name: "empty option leaves CRLF", in: "a\r\nb", option: "", want: "a\r\nb"},
		{name: "empty option leaves CR", in: "a\rb", option: "", want: "a\rb"},

		{name: "lf CRLF->LF", in: "a\r\nb", option: "lf", want: "a\nb"},
		{name: "lf CR->LF", in: "a\rb", option: "lf", want: "a\nb"},
		{name: "lf mixed", in: "a\r\nb\rc\nd", option: "lf", want: "a\nb\nc\nd"},
		{name: "lf already lf", in: "a\nb", option: "lf", want: "a\nb"},

		{name: "crlf LF->CRLF", in: "a\nb", option: "crlf", want: "a\r\nb"},
		{name: "crlf CR->CRLF", in: "a\rb", option: "crlf", want: "a\r\nb"},
		{name: "crlf already crlf", in: "a\r\nb", option: "crlf", want: "a\r\nb"},
		{name: "crlf leading lf", in: "\nb", option: "crlf", want: "\r\nb"},
		{name: "crlf trailing lf", in: "a\n", option: "crlf", want: "a\r\n"},
		{name: "crlf mixed", in: "a\r\nb\rc\nd", option: "crlf", want: "a\r\nb\r\nc\r\nd"},

		{name: "LF uppercase", in: "a\r\nb", option: "LF", want: "a\nb"},
		{name: "CRLF uppercase", in: "a\nb", option: "CRLF", want: "a\r\nb"},

		{name: "unknown option unchanged", in: "a\nb", option: "foo", want: "a\nb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertLineEnding(tt.in, tt.option)
			if got != tt.want {
				t.Errorf("ConvertLineEnding(%q, %q) = %q, want %q", tt.in, tt.option, got, tt.want)
			}
		})
	}
}

func TestConvertLineEnding_RoundTrip(t *testing.T) {
	src := "line1\r\nline2\rline3\nline4"
	lf := ConvertLineEnding(src, "lf")
	crlf := ConvertLineEnding(lf, "crlf")
	if crlf != ConvertLineEnding(src, "crlf") {
		t.Errorf("idempotent normalization mismatch: %q vs %q", crlf, ConvertLineEnding(src, "crlf"))
	}
}
