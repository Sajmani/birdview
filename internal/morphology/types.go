package morphology

// BirdSpecies represents a single compiled avian record ready for client consumption.
type BirdSpecies struct {
	ID             string   `json:"id"`
	CommonName     string   `json:"commonName"`
	ScientificName string   `json:"scientificName"`
	Family         string   `json:"family"`
	FamilyCommon   string   `json:"familyCommon"`
	FamilyColor    string   `json:"familyColor"`
	Regions        []string `json:"regions"`

	// Macroscopic Dimensions
	LengthIn   float64 `json:"lengthIn"`
	WingspanIn float64 `json:"wingspanIn"`
	MassG      float64 `json:"massG"`
	TotalMm    float64 `json:"totalMm"`
	WingspanMm float64 `json:"wingspanMm"`

	// Internal Segmented Coordinates (mm)
	CulmenMm float64 `json:"culmenMm"`
	HeadMm   float64 `json:"headMm"`
	NeckMm   float64 `json:"neckMm"`
	TorsoMm  float64 `json:"torsoMm"`
	TailMm   float64 `json:"tailMm"`
}
