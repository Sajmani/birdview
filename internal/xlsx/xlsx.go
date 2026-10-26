// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package xlsx reads the cell text of an Excel workbook, row by row.
//
// It does only what BirdView's three source readers need: list the sheets,
// and stream one sheet's rows as strings placed by column. Numbers come back
// as the text stored in the file, unformatted; booleans as TRUE or FALSE.
// Styles, formulas, dates, and writing are out of scope. A workbook is a zip
// of XML parts, so the standard library is enough, and the vendored sources
// are read without a third-party parser.
package xlsx

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// File is an open workbook.
type File struct {
	zr      *zip.ReadCloser
	sheets  []sheet
	strings []string
}

type sheet struct{ name, part string }

// Open opens the workbook at path and reads its sheet list and shared strings.
func Open(name string) (*File, error) {
	zr, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	f := &File{zr: zr}
	if err := f.readSheets(); err != nil {
		zr.Close()
		return nil, fmt.Errorf("xlsx %s: %w", name, err)
	}
	if err := f.readSharedStrings(); err != nil {
		zr.Close()
		return nil, fmt.Errorf("xlsx %s: %w", name, err)
	}
	return f, nil
}

// Close releases the file.
func (f *File) Close() error { return f.zr.Close() }

// SheetNames returns the sheet names in workbook order.
func (f *File) SheetNames() []string {
	out := make([]string, len(f.sheets))
	for i, s := range f.sheets {
		out[i] = s.name
	}
	return out
}

func (f *File) open(part string) (io.ReadCloser, error) {
	for _, zf := range f.zr.File {
		if zf.Name == part {
			return zf.Open()
		}
	}
	return nil, fmt.Errorf("no part %s", part)
}

// readSheets joins workbook.xml's sheet names to their parts through the
// workbook's relationships.
func (f *File) readSheets() error {
	targets := map[string]string{}
	err := f.each("xl/_rels/workbook.xml.rels", func(d *xml.Decoder, se xml.StartElement) error {
		if se.Name.Local == "Relationship" {
			targets[attr(se, "Id")] = attr(se, "Target")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return f.each("xl/workbook.xml", func(d *xml.Decoder, se xml.StartElement) error {
		if se.Name.Local != "sheet" {
			return nil
		}
		var id string
		for _, a := range se.Attr {
			if a.Name.Local == "id" && strings.HasSuffix(a.Name.Space, "/relationships") {
				id = a.Value
			}
		}
		target, ok := targets[id]
		if !ok {
			return fmt.Errorf("sheet %q has no relationship %q", attr(se, "name"), id)
		}
		part := path.Clean("xl/" + target)
		if strings.HasPrefix(target, "/") {
			part = path.Clean(target[1:])
		}
		f.sheets = append(f.sheets, sheet{attr(se, "name"), part})
		return nil
	})
}

func (f *File) readSharedStrings() error {
	rc, err := f.open("xl/sharedStrings.xml")
	if err != nil {
		return nil // a workbook with no text cells has no shared strings
	}
	defer rc.Close()
	d := xml.NewDecoder(rc)
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "si" {
			s, err := text(d, "si")
			if err != nil {
				return err
			}
			f.strings = append(f.strings, s)
		}
	}
}

// each calls fn for every start element in a part.
func (f *File) each(part string, fn func(*xml.Decoder, xml.StartElement) error) error {
	rc, err := f.open(part)
	if err != nil {
		return err
	}
	defer rc.Close()
	d := xml.NewDecoder(rc)
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if se, ok := tok.(xml.StartElement); ok {
			if err := fn(d, se); err != nil {
				return err
			}
		}
	}
}

// Rows streams the rows of the named sheet.
func (f *File) Rows(name string) (*Rows, error) {
	for _, s := range f.sheets {
		if s.name == name {
			rc, err := f.open(s.part)
			if err != nil {
				return nil, err
			}
			return &Rows{f: f, rc: rc, d: xml.NewDecoder(rc)}, nil
		}
	}
	return nil, fmt.Errorf("no sheet named %q", name)
}

// Rows iterates over a sheet's rows. Call Next before each Columns.
type Rows struct {
	f   *File
	rc  io.ReadCloser
	d   *xml.Decoder
	cur []string
	err error
}

// Next advances to the next row, reporting whether there is one.
func (r *Rows) Next() bool {
	r.cur = nil
	for {
		tok, err := r.d.Token()
		if err != nil {
			if err != io.EOF {
				r.err = err
			}
			return false
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "row" {
			r.cur, r.err = r.row()
			return r.err == nil
		}
	}
}

// Columns returns the current row's cells, each at its column's index; a
// column with no cell is the empty string.
func (r *Rows) Columns() ([]string, error) { return r.cur, r.err }

// Close releases the sheet.
func (r *Rows) Close() error { return r.rc.Close() }

func (r *Rows) row() ([]string, error) {
	var cells []string
	next := 0
	for {
		tok, err := r.d.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local != "c" {
				continue
			}
			col := next
			if ref := attr(t, "r"); ref != "" {
				if col, err = column(ref); err != nil {
					return nil, err
				}
			}
			v, err := r.cell(t)
			if err != nil {
				return nil, err
			}
			for len(cells) < col {
				cells = append(cells, "")
			}
			cells = append(cells, v)
			next = col + 1
		case xml.EndElement:
			if t.Name.Local == "row" {
				return cells, nil
			}
		}
	}
}

// cell reads one <c> element and returns its text.
func (r *Rows) cell(c xml.StartElement) (string, error) {
	var v, inline string
	for {
		tok, err := r.d.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "v":
				if v, err = text(r.d, "v"); err != nil {
					return "", err
				}
			case "is":
				if inline, err = text(r.d, "is"); err != nil {
					return "", err
				}
			}
		case xml.EndElement:
			if t.Name.Local != "c" {
				continue
			}
			switch attr(c, "t") {
			case "s":
				i, err := strconv.Atoi(v)
				if err != nil || i < 0 || i >= len(r.f.strings) {
					return "", fmt.Errorf("cell %s: bad shared string %q", attr(c, "r"), v)
				}
				return r.f.strings[i], nil
			case "inlineStr":
				return inline, nil
			case "b":
				if v == "1" {
					return "TRUE", nil
				}
				return "FALSE", nil
			}
			return v, nil
		}
	}
}

// text reads the element just opened through its end tag named end. A <v>
// is its own text; anything else is the concatenation of its <t> runs, so
// whitespace between runs is not text. Phonetic runs (<rPh>) are not part
// of the cell's text.
func text(d *xml.Decoder, end string) (string, error) {
	var b strings.Builder
	depth, inT, phonetic := 0, 0, 0
	for {
		tok, err := d.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "t":
				inT++
			case "rPh":
				phonetic++
			}
		case xml.EndElement:
			if depth == 0 && t.Name.Local == end {
				return b.String(), nil
			}
			depth--
			switch t.Name.Local {
			case "t":
				inT--
			case "rPh":
				phonetic--
			}
		case xml.CharData:
			if phonetic == 0 && (inT > 0 || end == "v") {
				b.Write(t)
			}
		}
	}
}

// column converts a cell reference such as "AB12" to a zero-based column.
func column(ref string) (int, error) {
	n := 0
	for i, ch := range ref {
		if ch >= 'A' && ch <= 'Z' {
			n = n*26 + int(ch-'A'+1)
			continue
		}
		if i == 0 {
			break
		}
		return n - 1, nil
	}
	return 0, fmt.Errorf("bad cell reference %q", ref)
}

func attr(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local && a.Name.Space == "" {
			return a.Value
		}
	}
	return ""
}
