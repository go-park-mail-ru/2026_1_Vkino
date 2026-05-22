package subscription

type AdPolicy string

const (
	AdPolicyNoSkip      AdPolicy = "no_skip"
	AdPolicySkipPreroll AdPolicy = "skip_preroll"
	AdPolicySkipAll     AdPolicy = "skip_all"
	AdPolicyNone        AdPolicy = "none"
)
