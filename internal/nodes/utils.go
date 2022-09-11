package nodes

func GenerateToasterAppsSubDomains(radix string) []string {
	var d []string

	RegionsMu.RLock()
	for i := 0; i < len(Regions); i++ {
		d = append(d, "*."+Regions[i]+"."+radix)
	}
	RegionsMu.RUnlock()

	return d
}

func GetToasterLocalRegionAppSubdomain(radix, localregion string) string {
	return "*." + localregion + "." + radix
}
