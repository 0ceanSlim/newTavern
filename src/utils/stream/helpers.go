package stream

import (
	"fmt"
	"math/rand"
)

func generateDtag() string {
	return fmt.Sprintf("%d", rand.Intn(900000)+100000)
}
