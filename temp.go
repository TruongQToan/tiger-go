package tiger

import "strconv"

type Temp uint64
type Label Symbol

var temp Temp = 100

func newTemp() Temp {
	temp++
	return temp
}

func makeTempString() string {
	return "t" + strconv.FormatInt(int64(temp), 10)
}
