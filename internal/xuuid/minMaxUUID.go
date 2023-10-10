package xuuid

import "github.com/google/uuid"

// all 1 value of uuid
var MAX = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

// all 0 value of uuid
var MIN = uuid.Nil
