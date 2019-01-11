package fizzbuzz

import (
	"time"
	"strconv"
	"github.com/wizgrao/blow/maps"
)

type fizzBuzz struct{
	Number int
	Word   string
}

type FizzMapper int

func (FizzMapper) Do(k maps.Keyed, c chan <- maps.Keyed) {
	n := k.(*fizzyinput).Val
	time.Sleep(time.Second/4)
	if n % 15 == 0 {
		c <- &fizzBuzz{
			Number: n,
			Word:   "fizzbuzz",
		}

	} else if n % 3 == 0 {
		c <- &fizzBuzz{
			Number: n,
			Word:   "fizz",
		}
	} else if n % 5 == 0 {
		c <- &fizzBuzz{
			Number: n,
			Word:   "buzz",
		}
	} else {
		c <- &fizzBuzz{
			Number: n,
			Word:   strconv.FormatInt(int64(n), 10),
		}
	}
}

func (f *fizzBuzz) Key() int {
	return f.Number
}

type FizzGenerator int


type fizzyinput struct {
	Val int
}

func (f *fizzyinput) Key() int {
	return f.Val
}
func (FizzGenerator) Do(c chan <- maps.Keyed) {
	for i:=0; i< 10000; i++ {
		c <- &fizzyinput{i}
	}
}
func (FizzMapper) ID() string{
	return "fizzymapper"
}
func (FizzMapper) NewIn() maps.Keyed {
	return &fizzyinput{}
}

func (FizzMapper) NewOut() maps.Keyed {
	return &fizzBuzz{}
}
