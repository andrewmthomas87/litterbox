package models

type Stage int

const (
	StageInformation Stage = iota + 1
	StageDefault
)

type Building int

const (
	Building1838Chicago Building = iota + 1
	Building2303SheridanGREEN
	Building2303Sheridan
	Building2349Sheridan
	Building560Lincoln
	Building720Emerson
	BuildingAllison
	BuildingAyers
	BuildingBobb
	BuildingChapin
	BuildingEastFairchild
	BuildingElder
	BuildingFosterWalker
	BuildingGoodrich
	BuildingHobart
	BuildingJones
	BuildingKemper
	BuildingMcCulloch
	BuildingNorthMidQuads
	BuildingRogers
	BuildingSargent
	BuildingShepard
	BuildingSlivka
	BuildingSouthMidQuads
	BuildingWestFairchild
	BuildingWillard
)

var BuildingLookup = map[string]Building{
	"1838 Chicago Ave.":                Building1838Chicago,
	"2303 Sheridan Road (GREEN House)": Building2303SheridanGREEN,
	"2303 Sheridan Road (Residential College of Cultural & Community Studies)": Building2303Sheridan,
	"2349 Sheridan Rd":                   Building2349Sheridan,
	"560 Lincoln":                        Building560Lincoln,
	"720 Emerson St. (Sigma Alpha Iota)": Building720Emerson,
	"Allison Hall":                       BuildingAllison,
	"Ayers Hall (Residential College of Commerce & Industry)": BuildingAyers,
	"Bobb": BuildingBobb,
	"Chapin Hall (Humanities Residential College)":        BuildingChapin,
	"East Fairchild (Communications Residential College)": BuildingEastFairchild,
	"Elder Hall":            BuildingElder,
	"Foster-Walker Complex": BuildingFosterWalker,
	"Goodrich House":        BuildingGoodrich,
	"Hobart House (Women's Residential College)": BuildingHobart,
	"Jones Hall":  BuildingJones,
	"Kemper Hall": BuildingKemper,
	"McCulloch":   BuildingMcCulloch,
	"North Mid-Quads (Public Affairs Residential College)": BuildingNorthMidQuads,
	"Rogers House": BuildingRogers,
	"Sargent Hall": BuildingSargent,
	"Shepard Hall": BuildingShepard,
	"Slivka Hall (Residential College of Science and Engineering)": BuildingSlivka,
	"South Mid-Quads (Shepard Residential College)":                BuildingSouthMidQuads,
	"West Fairchild (International Studies Residential College)":   BuildingWestFairchild,
	"Willard Hall (Willard Residential College)":                   BuildingWillard,
}

type InformationErrors struct {
	Name     string `json:"name"`
	Building string `json:"building"`
	Address  string `json:"address"`
}
