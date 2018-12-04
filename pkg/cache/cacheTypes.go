package cache

import (
	"time"
)

//cache object "types"
const (
	CacheTypeResource             = "static:"
	CacheTypePlugin               = "Plugin:"
	CacheTypeDatabase             = "Database:"
	CacheTypeTemplateHelperPlugin = "templateHelperPlugin:"
)

//MaxTTL is the max possible TTL value (aprox 290 years)
const MaxTTL = time.Duration(^(uint64(1) << 63))
