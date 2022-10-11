package features

type FeatureMap struct {
	Features []*Feature `json:"features"`
}
var featureMap *FeatureMap

type Feature struct {
	Name   string `json:"name"`
	Toggle bool   `json:"toggle"`
}

func SetFeatures(features []*Feature) {
	if featureMap == nil {
		featureMap = &FeatureMap{}
	}
	featureMap.Features = features
}
func IsEnabled(feature string) bool {
	for _, f := range featureMap.Features {
		if feature == f.Name {
			return f.Toggle
		}
	}
	return false
}
