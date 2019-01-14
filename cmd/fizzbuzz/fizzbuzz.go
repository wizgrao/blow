package fizzbuzz

import (
	"time"
	"strconv"
	"github.com/wizgrao/blow/maps"
	"github.com/golang/protobuf/proto"
)


type FizzGenerator int

func (FizzGenerator) Do(c chan <- maps.Keyed) {
	for i:=0; i< 10000; i++ {
		c <- &FizzyInput{Val: int32(i)}
	}
}

type FizzMapper int

func (FizzMapper) Do(k maps.Keyed, c chan <- maps.Keyed) {
	n := k.(*FizzyInput).Val
	time.Sleep(time.Second/4)
	if n % 15 == 0 {
		c <- &FizzBuzz{
			Number: n,
			Word:   "fizzbuzz",
		}

	} else if n % 3 == 0 {
		c <- &FizzBuzz{
			Number: n,
			Word:   "fizz",
		}
	} else if n % 5 == 0 {
		c <- &FizzBuzz{
			Number: n,
			Word:   "buzz",
		}
	} else {
		c <- &FizzBuzz{
			Number: n,
			Word:   strconv.FormatInt(int64(n), 10),
		}
	}
}

func (FizzMapper) ID() string{
	return "fizzymapper"
}

func (f FizzMapper) InEncoder() maps.Encoder {
	return &FizzyInputMarshaller{}
}
func (f FizzMapper) OutEncoder() maps.Encoder {
	return &FizzBuzzMarshaller{}
}

func (f *FizzyInput) Key() int {
	return int(f.Val)
}

func (f *FizzBuzz) Key() int {
	return int(f.Number)
}

type FizzyInputMarshaller struct {

}

func (*FizzyInputMarshaller) Marshal(k maps.Keyed) ([]byte, error) {
	return proto.Marshal(k.(*FizzyInput))
}
func (*FizzyInputMarshaller) UnMarshal(b []byte) (maps.Keyed, error) {
	o := &FizzyInput{}
	err := proto.Unmarshal(b, o)
	return o, err
}

type FizzBuzzMarshaller struct {

}

func (*FizzBuzzMarshaller) Marshal(k maps.Keyed) ([]byte, error) {
	return proto.Marshal(k.(*FizzBuzz))
}
func (*FizzBuzzMarshaller) UnMarshal(b []byte) (maps.Keyed, error) {
	o := &FizzBuzz{}
	err := proto.Unmarshal(b, o)
	return o, err
}

