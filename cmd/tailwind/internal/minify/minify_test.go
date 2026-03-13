package minify

import "testing"

func TestCSS(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "remove comments",
			input: "/* comment */ .a { color: red; }",
			want:  ".a{color:red}",
		},
		{
			name:  "collapse whitespace",
			input: ".a  {  color:  red;  }",
			want:  ".a{color:red}",
		},
		{
			name:  "remove trailing semicolons",
			input: ".a { color: red; }",
			want:  ".a{color:red}",
		},
		{
			name:  "preserve double-quoted strings",
			input: `.a { content: "  hello  "; }`,
			want:  `.a{content:"  hello  "}`,
		},
		{
			name:  "preserve single-quoted strings",
			input: ".a { font-family: 'Times New Roman'; }",
			want:  ".a{font-family:'Times New Roman'}",
		},
		{
			name:  "nested braces media queries",
			input: "@media (min-width: 768px) { .a { color: red; } }",
			want:  "@media(min-width:768px){.a{color:red}}",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "already minified",
			input: ".a{color:red}",
			want:  ".a{color:red}",
		},
		{
			name:  "multiline CSS",
			input: ".container {\n  display: flex;\n  justify-content: center;\n  align-items: center;\n}\n\n.text-red {\n  color: red;\n}\n",
			want:  ".container{display:flex;justify-content:center;align-items:center}.text-red{color:red}",
		},
		{
			name:  "url values preserved",
			input: "background: url( image.png )",
			want:  "background:url(image.png)",
		},
		{
			name:  "url with quotes preserved",
			input: `background: url("image.png")`,
			want:  `background:url("image.png")`,
		},
		{
			name:  "multiple comments",
			input: "/* a */ .a { /* b */ color: red; /* c */ }",
			want:  ".a{color:red}",
		},
		{
			name:  "multiple selectors with comma",
			input: ".a, .b { color: red; }",
			want:  ".a,.b{color:red}",
		},
		{
			name:  "multiple declarations",
			input: ".a { color: red; background: blue; }",
			want:  ".a{color:red;background:blue}",
		},
		{
			name:  "escaped characters in strings",
			input: `.a { content: "hello \"world\""; }`,
			want:  `.a{content:"hello \"world\""}`,
		},
		{
			name:  "whitespace only",
			input: "   \n\t  ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CSS(tt.input)
			if got != tt.want {
				t.Errorf("CSS(%q)\n got  %q\n want %q", tt.input, got, tt.want)
			}
		})
	}
}
