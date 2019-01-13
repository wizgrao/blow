package fizzbuzz

import (
	"time"
	"strconv"
	"github.com/wizgrao/blow/maps"
	"encoding/json"
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

func (f FizzMapper) InEncoder() maps.Encoder {
	return &FizzyInputMarshaller{}
}
func (f FizzMapper) OutEncoder() maps.Encoder {
	return &FizzBuzzMarshaller{}
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
type FizzyInputMarshaller struct {

}

func (*FizzyInputMarshaller) Marshal(k maps.Keyed) ([]byte, error) {
	return json.Marshal(k)
}
func (*FizzyInputMarshaller) UnMarshal(b []byte) (maps.Keyed, error) {
	o := &fizzyinput{}
	err := json.Unmarshal(b, o)
	return o, err
}

type FizzBuzzMarshaller struct {

}

func (*FizzBuzzMarshaller) Marshal(k maps.Keyed) ([]byte, error) {
	return json.Marshal(k)
}
func (*FizzBuzzMarshaller) UnMarshal(b []byte) (maps.Keyed, error) {
	o := &fizzBuzz{}
	err := json.Unmarshal(b, o)
	return o, err
}

