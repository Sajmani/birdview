package wingspan

import (
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// prediction is one of Fox's own published estimates, used only to validate
// that BirdView's fitted model reconstructs theirs.
type prediction struct {
	species, order string
	chordMm, foxCm float64
}

func publishedPredictions(path string) ([]prediction, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := f.Rows("wingspan data")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var idx map[string]int
	var out []prediction
	for rows.Next() {
		c, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		if idx == nil {
			idx = map[string]int{}
			for i, h := range c {
				idx[strings.TrimSpace(h)] = i
			}
			continue
		}
		get := func(k string) string { return at(c, idx[k]) }
		if isEmpirical(get("needs_prediction")) {
			continue
		}
		fox := num(get("wingspan_predicted_order"))
		if fox == 0 {
			fox = num(get("wingspan_predicted_global"))
		}
		chord := num(get("Wing.Length_AVONET"))
		if fox <= 0 || chord <= 0 {
			continue
		}
		out = append(out, prediction{
			species: strings.TrimSpace(get("Species1")),
			order:   strings.TrimSpace(get("Order1")),
			chordMm: chord, foxCm: fox,
		})
	}
	return out, nil
}

var _ = strconv.Atoi
