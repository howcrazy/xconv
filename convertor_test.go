package goconv

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestConvertor(t *testing.T) {
	{
		var src int = 100
		var dst int32
		Convert(src, &dst)
		debug(src, dst)
	}
	{
		src := [3]int32{1, 2, 3}
		var dst []int
		Convert(src, &dst)
		debug(src, dst)
	}
	{
		src1 := time.Now()
		var (
			dst1 string
			dst2 time.Time
			dst3 int32
			dst4 time.Time
		)
		Convert(src1, &dst1)
		debug(src1, dst1)

		Convert(dst1, &dst2)
		debug("\n0:\t%s\n1:\t%s", dst1, dst2)

		Convert(src1, &dst3)
		debug(src1, dst3, src1.Unix())

		Convert(dst3, &dst4)
		debug(dst3, dst4)
	}
	{
		type Src struct {
			A int
			B string
		}
		type Dst struct {
			A int32
			B string
		}
		src := []*Src{&Src{A: 1, B: "one"}, &Src{A: 2, B: "two"}}
		var dst []Dst
		Convert(src, &dst)
		debug(src, dst)
	}
	{
		var dst1 byte
		Convert(byte('a'), &dst1)
		debug(dst1)

		var dst2 []byte
		Convert([]byte("abc"), &dst2)
		debug(dst2)
	}
	{
		src := map[int]string{1: "one", 2: "tow"}
		var dst map[int32]string
		Convert(src, &dst)
		debug(src, dst)
	}
}

func TestConvertor1(t *testing.T) {
	{
		type Base struct {
			B int
		}
		type Src struct {
			A    string
			B    int
			C    time.Time
			D    int32
			x    int
			Z    []int
			Same []*Base
		}
		type Dst struct {
			*Base
			A    string
			C    string
			x    int
			E    string
			Z    []int32
			Same []*Base
		}
		dst := Dst{}
		src := &Src{A: "one", B: 1024, C: time.Now(), D: 32, Z: []int{1, 2, 3}, Same: []*Base{&Base{B: 10}}}
		NewConvertor(src).
			Rule(reflect.String, reflect.String,
			func(c *Convertor, src, dst reflect.Value) {
				dst.SetString(src.String() + "xxxxxxxxxxxxx")
			}).
			Field("E",
			func(srcObj *Src) string {
				return strconv.Itoa(int(srcObj.D + 100))
			}).
			Field("Base",
			func(srcObj *Src) *Base {
				return &Base{B: srcObj.B}
			}).
			TimeFormat(DATE_FORMAT).
			Apply(&dst)
		dst.Same[0].B = 1000
		debug(src, dst)
	}
	{
		type Src struct {
			A string
		}
		type Dst struct {
			B string
		}
		src1 := map[int]Src{1: Src{"one"}, 2: Src{"two"}}
		var dst1 map[int32]*Dst
		NewConvertor(src1).Field("B", func(obj Src) string { return obj.A }).Apply(&dst1)
		debug(src1, dst1)

		src2 := map[int]*Src{1: &Src{"one"}, 2: &Src{"two"}}
		var dst2 map[int32]Dst
		NewConvertor(src2).Field("B", func(obj *Src) string { return obj.A }).Apply(&dst2)
		debug(src2, dst2)
	}
	{
		type SrcBase struct {
			A string
		}
		type Src struct {
			Base SrcBase
			A    string
		}
		type DstBase struct {
			B int
		}
		type Dst struct {
			Base *DstBase
			B    string
			DstBase
		}
		src := Src{Base: SrcBase{"one"}, A: "two"}
		var dst *Dst
		dst = new(Dst)
		NewConvertor(src).
			Field("B", func(obj Src) string { return obj.A }).
			Field("Base.B", func(obj SrcBase) int { return 1 }).
			Field("DstBase", func(obj Src) DstBase { return DstBase{B: 3} }).
			Apply(dst)
		debug(src, dst)
	}
}
