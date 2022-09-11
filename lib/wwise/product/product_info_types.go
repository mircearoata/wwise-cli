package product

type ProductInfo struct {
	Bundles []Bundle `json:"bundles"`
}
type PlatformFolders struct {
	Mandatory []string `json:"mandatory"`
	Optional  []string `json:"optional"`
}
type SupportedUnrealVersions struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}
type ProductDependentData struct {
	PlatformFolders         PlatformFolders           `json:"platformFolders"`
	SupportedPlatforms      []string                  `json:"supportedPlatforms"`
	SupportedUnrealVersions []SupportedUnrealVersions `json:"supportedUnrealVersions"`
	WwiseSdkBuild           int                       `json:"wwiseSdkBuild"`
}
type Version struct {
	Build    int    `json:"build"`
	Major    int    `json:"major"`
	Minor    int    `json:"minor"`
	Nickname string `json:"nickname"`
	Year     int    `json:"year"`
}
type MinimumRequiredVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Year  int `json:"year"`
}
type Launcher struct {
	MinimumRequiredVersion MinimumRequiredVersion `json:"minimumRequiredVersion"`
}
type Bundle struct {
	ID                   string               `json:"id"`
	Tag                  string               `json:"tag"`
	VersionTag           string               `json:"versionTag"`
	Published            int                  `json:"published"`
	Stable               int                  `json:"stable"`
	Type                 string               `json:"type"`
	Name                 string               `json:"name"`
	VersionName          string               `json:"versionName"`
	Description          interface{}          `json:"description"`
	RequiredLicenseID    int                  `json:"requiredLicenseId"`
	ProductDependentData ProductDependentData `json:"productDependentData"`
	Vendor               string               `json:"vendor"`
	Version              Version              `json:"version"`
	Documentation        []interface{}        `json:"documentation"`
	Links                []interface{}        `json:"links"`
	SortGroupIndex       int                  `json:"$sortGroupIndex"`
	AutomaticDeletion    int                  `json:"automaticDeletion"`
	Supported            int                  `json:"supported"`
	CategoryID           int                  `json:"categoryId"`
	Image                interface{}          `json:"image"`
	Labels               []interface{}        `json:"labels"`
	Launcher             Launcher             `json:"launcher,omitempty"`
}

type ProductVersionInfo struct {
	Eulas                []Eula               `json:"eulas"`
	Files                []File               `json:"files"`
	Groups               []GroupList          `json:"groups"`
	ID                   string               `json:"id"`
	Labels               []interface{}        `json:"labels"`
	Launcher             Launcher             `json:"launcher"`
	Links                []interface{}        `json:"links"`
	Name                 string               `json:"name"`
	ProductDependentData ProductDependentData `json:"productDependentData"`
	Tag                  string               `json:"tag"`
	Type                 string               `json:"type"`
	Vendor               string               `json:"vendor"`
	Version              Version              `json:"version"`
	SortGroupIndex       int                  `json:"$sortGroupIndex"`
}
type Eula struct {
	DisplayName string `json:"displayName"`
	FileName    string `json:"fileName"`
	ID          string `json:"id"`
}
type Group struct {
	GroupID      string `json:"groupId"`
	GroupValueID string `json:"groupValueId"`
}
type File struct {
	DocumentationFiles []interface{} `json:"documentationFiles"`
	Groups             []Group       `json:"groups"`
	ID                 string        `json:"id"`
	Licenses           []interface{} `json:"licenses"`
	Name               string        `json:"name"`
	Sha1               string        `json:"sha1"`
	Size               int           `json:"size"`
	SourceName         string        `json:"sourceName"`
	UncompressedSize   int           `json:"uncompressedSize"`
	URL                string        `json:"url"`
}
type License struct {
	Platform string `json:"platform"`
}
type Values struct {
	Description string   `json:"description"`
	DisplayName string   `json:"displayName"`
	EulaIds     []string `json:"eulaIds"`
	ID          string   `json:"id"`
	License     License  `json:"license"`
}
type GroupList struct {
	DisplayName string   `json:"displayName"`
	ID          string   `json:"id"`
	Values      []Values `json:"values"`
}
