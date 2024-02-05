package http_rnd

import (
	"fmt"
	"math/rand"
	"time"
)

const rndNum = 2756983592

func MakeReqId() (string, error) {
	const op = "pkg.random.makeReqId"
	b := make([]byte, 16)

	s := rand.NewSource(time.Now().Unix() + rndNum)
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	return fmt.Sprintf("%x", b), nil
}
