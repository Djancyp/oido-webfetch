package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type WebPage struct {
	URL         string
	Title       string
	Description string
	H1          []string
	H2          []string
	H3          []string
	H4          []string
	H5          []string
	H6          []string
	Paragraphs  []string
	Links       []PageLink
	Buttons     []string
	Images      []PageImage
	Lists       []PageList
	Tables      []PageTable
	Forms       []PageForm
	CodeBlocks  []string
}

type PageLink struct {
	Text string
	Href string
}

type PageImage struct {
	Src string
	Alt string
}

type PageList struct {
	Items []string
}

type PageTable struct {
	Headers []string
	Rows    [][]string
}

type PageForm struct {
	Action string
	Method string
	Inputs []FormInput
}

type FormInput struct {
	Name        string
	Type        string
	Placeholder string
	Value       string
}

func extractPageStructure(doc *goquery.Document) WebPage {
	page := WebPage{}

	page.Title = strings.TrimSpace(doc.Find("title").First().Text())

	page.Description = strings.TrimSpace(doc.Find(`meta[name="description"]`).AttrOr("content", ""))

	doc.Find("h1").Each(func(_ int, s *goquery.Selection) {
		page.H1 = append(page.H1, trimmedNonEmpty(s.Text()))
	})

	doc.Find("h2").Each(func(_ int, s *goquery.Selection) {
		page.H2 = append(page.H2, trimmedNonEmpty(s.Text()))
	})

	doc.Find("h3").Each(func(_ int, s *goquery.Selection) {
		page.H3 = append(page.H3, trimmedNonEmpty(s.Text()))
	})

	doc.Find("h4").Each(func(_ int, s *goquery.Selection) {
		page.H4 = append(page.H4, trimmedNonEmpty(s.Text()))
	})

	doc.Find("h5").Each(func(_ int, s *goquery.Selection) {
		page.H5 = append(page.H5, trimmedNonEmpty(s.Text()))
	})

	doc.Find("h6").Each(func(_ int, s *goquery.Selection) {
		page.H6 = append(page.H6, trimmedNonEmpty(s.Text()))
	})

	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		t := trimmedNonEmpty(s.Text())
		if t != "" {
			page.Paragraphs = append(page.Paragraphs, t)
		}
	})

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		text := trimmedNonEmpty(s.Text())
		if href != "" && !strings.HasPrefix(href, "#") {
			page.Links = append(page.Links, PageLink{Text: text, Href: href})
		}
	})

	doc.Find("button, input[type=submit], input[type=button]").Each(func(_ int, s *goquery.Selection) {
		t := trimmedNonEmpty(s.Text())
		if t == "" {
			if v, ok := s.Attr("value"); ok {
				t = trimmedNonEmpty(v)
			}
		}
		if t != "" {
			page.Buttons = append(page.Buttons, t)
		}
	})

	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		alt, _ := s.Attr("alt")
		if src != "" {
			page.Images = append(page.Images, PageImage{Src: src, Alt: alt})
		}
	})

	doc.Find("ul, ol").Each(func(_ int, s *goquery.Selection) {
		var items []string
		s.Find("li").Each(func(_ int, li *goquery.Selection) {
			t := trimmedNonEmpty(li.Text())
			if t != "" {
				items = append(items, t)
			}
		})
		if len(items) > 0 {
			page.Lists = append(page.Lists, PageList{Items: items})
		}
	})

	doc.Find("table").Each(func(_ int, s *goquery.Selection) {
		var headers []string
		s.Find("thead tr th, tr th").Each(func(_ int, th *goquery.Selection) {
			headers = append(headers, trimmedNonEmpty(th.Text()))
		})
		var rows [][]string
		s.Find("tbody tr, tr").Each(func(_ int, tr *goquery.Selection) {
			var cells []string
			tr.Find("td").Each(func(_ int, td *goquery.Selection) {
				cells = append(cells, trimmedNonEmpty(td.Text()))
			})
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
		})
		if len(rows) > 0 || len(headers) > 0 {
			page.Tables = append(page.Tables, PageTable{Headers: headers, Rows: rows})
		}
	})

	doc.Find("form").Each(func(_ int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		method, _ := s.Attr("method")
		if method == "" {
			method = "GET"
		}
		var inputs []FormInput
		s.Find("input, select, textarea").Each(func(_ int, inp *goquery.Selection) {
			name, _ := inp.Attr("name")
			typ, _ := inp.Attr("type")
			if typ == "" {
				typ = "text"
			}
			placeholder, _ := inp.Attr("placeholder")
			value, _ := inp.Attr("value")

			if tag := goquery.NodeName(inp); tag == "select" {
				inp.Find("option[selected]").Each(func(_ int, opt *goquery.Selection) {
					if v, ok := opt.Attr("value"); ok {
						value = v
					}
				})
			}

			if name != "" || typ != "" {
				inputs = append(inputs, FormInput{
					Name:        name,
					Type:        typ,
					Placeholder: placeholder,
					Value:       value,
				})
			}
		})
		if action != "" || len(inputs) > 0 {
			page.Forms = append(page.Forms, PageForm{
				Action: action,
				Method: method,
				Inputs: inputs,
			})
		}
	})

	doc.Find("pre, code").Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" && len(t) > 10 {
			page.CodeBlocks = append(page.CodeBlocks, t)
		}
	})

	return page
}

func formatAsYAML(page WebPage) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("url: %s\n", page.URL))
	if page.Title != "" {
		sb.WriteString(fmt.Sprintf("title: %q\n", page.Title))
	}
	if page.Description != "" {
		sb.WriteString(fmt.Sprintf("description: %q\n", page.Description))
	}

	writeYAMLList(&sb, "h1", page.H1)
	writeYAMLList(&sb, "h2", page.H2)
	writeYAMLList(&sb, "h3", page.H3)
	writeYAMLList(&sb, "h4", page.H4)
	writeYAMLList(&sb, "h5", page.H5)
	writeYAMLList(&sb, "h6", page.H6)

	if len(page.Paragraphs) > 0 {
		sb.WriteString("paragraphs:\n")
		for _, p := range page.Paragraphs {
			if len(p) > 300 {
				p = p[:300] + "..."
			}
			sb.WriteString(fmt.Sprintf("  - %q\n", p))
		}
	}

	if len(page.Links) > 0 {
		sb.WriteString("links:\n")
		for _, l := range page.Links {
			sb.WriteString(fmt.Sprintf("  - text: %q\n", l.Text))
			sb.WriteString(fmt.Sprintf("    href: %q\n", l.Href))
		}
	}

	if len(page.Buttons) > 0 {
		writeYAMLList(&sb, "buttons", page.Buttons)
	}

	if len(page.Images) > 0 {
		sb.WriteString("images:\n")
		for _, i := range page.Images {
			sb.WriteString(fmt.Sprintf("  - src: %q\n", i.Src))
			if i.Alt != "" {
				sb.WriteString(fmt.Sprintf("    alt: %q\n", i.Alt))
			}
		}
	}

	if len(page.Lists) > 0 {
		sb.WriteString("lists:\n")
		for _, l := range page.Lists {
			sb.WriteString("  - items:\n")
			for _, item := range l.Items {
				sb.WriteString(fmt.Sprintf("      - %q\n", item))
			}
		}
	}

	if len(page.Tables) > 0 {
		sb.WriteString("tables:\n")
		for _, t := range page.Tables {
			if len(t.Headers) > 0 {
				sb.WriteString("  - headers:\n")
				for _, h := range t.Headers {
					sb.WriteString(fmt.Sprintf("      - %q\n", h))
				}
			}
			if len(t.Rows) > 0 {
				sb.WriteString("    rows:\n")
				for _, row := range t.Rows {
					sb.WriteString("      -")
					for _, cell := range row {
						sb.WriteString(fmt.Sprintf(" %q", cell))
					}
					sb.WriteByte('\n')
				}
			}
		}
	}

	if len(page.Forms) > 0 {
		sb.WriteString("forms:\n")
		for _, f := range page.Forms {
			sb.WriteString(fmt.Sprintf("  - action: %q\n", f.Action))
			sb.WriteString(fmt.Sprintf("    method: %q\n", f.Method))
			if len(f.Inputs) > 0 {
				sb.WriteString("    inputs:\n")
				for _, inp := range f.Inputs {
					sb.WriteString(fmt.Sprintf("      - name: %q\n", inp.Name))
					sb.WriteString(fmt.Sprintf("        type: %q\n", inp.Type))
					if inp.Placeholder != "" {
						sb.WriteString(fmt.Sprintf("        placeholder: %q\n", inp.Placeholder))
					}
					if inp.Value != "" {
						sb.WriteString(fmt.Sprintf("        value: %q\n", inp.Value))
					}
				}
			}
		}
	}

	if len(page.CodeBlocks) > 0 {
		sb.WriteString("code_blocks:\n")
		for _, c := range page.CodeBlocks {
			lines := strings.Split(c, "\n")
			if len(lines) > 20 {
				lines = lines[:20]
				c = strings.Join(lines, "\n") + "\n  ..."
			}
			sb.WriteString("  - |\n")
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("      %s\n", line))
			}
		}
	}

	return sb.String()
}

func writeYAMLList(sb *strings.Builder, key string, items []string) {
	if len(items) == 0 {
		return
	}
	sb.WriteString(fmt.Sprintf("%s:\n", key))
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("  - %q\n", item))
	}
}

func trimmedNonEmpty(s string) string {
	t := strings.TrimSpace(s)
	return t
}

func formatSize(n int64) string {
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	case n >= 1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
