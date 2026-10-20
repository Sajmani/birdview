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

package xlsx

import (
	"archive/zip"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// workbook writes a minimal two-sheet workbook. The second sheet exercises
// every cell form the reader handles: a shared string with rich-text runs
// and a phonetic run, an inline string, a boolean, a number stored with more
// precision than any display format would show, and a skipped column.
func workbook(t *testing.T) string {
	t.Helper()
	parts := map[string]string{
		"xl/workbook.xml": `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>
 <sheet name="Notes" sheetId="1" r:id="rId1"/><sheet name="Data" sheetId="2" r:id="rId2"/>
 </sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
 <Relationship Id="rId1" Target="worksheets/sheet1.xml"/>
 <Relationship Id="rId2" Target="/xl/worksheets/sheet2.xml"/></Relationships>`,
		"xl/sharedStrings.xml": `<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
 <si><t>Species1</t></si>
 <si>
   <r><t>Larus </t></r>
   <r><rPr><i/></rPr><t>vegae</t></r>
   <rPh><t>IGNORED</t></rPh>
 </si></sst>`,
		"xl/worksheets/sheet1.xml": `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData/></worksheet>`,
		"xl/worksheets/sheet2.xml": `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
 <row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="inlineStr"><is><t>Mass</t></is></c></row>
 <row r="2"><c r="A2" t="s"><v>1</v></c><c r="B2"><v>108.16</v></c><c r="D2" t="b"><v>0</v></c></row>
 </sheetData></worksheet>`,
	}
	p := filepath.Join(t.TempDir(), "book.xlsx")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return p
}

func TestReadsCellsAsStored(t *testing.T) {
	f, err := Open(workbook(t))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if got, want := f.SheetNames(), []string{"Notes", "Data"}; !reflect.DeepEqual(got, want) {
		t.Errorf("SheetNames() = %q, want %q", got, want)
	}
	rows, err := f.Rows("Data")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got [][]string
	for rows.Next() {
		c, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, c)
	}
	want := [][]string{
		{"Species1", "Mass"},
		// 108.16 must not come back as 108.2: excelize returned values
		// rounded to the cell's display format, and BirdView's own reader
		// exists in part to read the stored value.
		{"Larus vegae", "108.16", "", "FALSE"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rows = %q, want %q", got, want)
	}
	if _, err := f.Rows("Missing"); err == nil {
		t.Error("Rows of a missing sheet succeeded")
	}
}

func TestColumn(t *testing.T) {
	for ref, want := range map[string]int{"A1": 0, "D2": 3, "Z9": 25, "AA10": 26, "AB12": 27} {
		if got, err := column(ref); err != nil || got != want {
			t.Errorf("column(%q) = %d, %v; want %d", ref, got, err, want)
		}
	}
	for _, bad := range []string{"", "12", "AB"} {
		if _, err := column(bad); err == nil {
			t.Errorf("column(%q) succeeded", bad)
		}
	}
}
