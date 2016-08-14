# XConv

[zh-CN](README_zh-cn.md)

XConv is a golang type convertor. It convert any value between types (base type,
 struct, array, slice, map, etc.)

## Features

- Convert between integer types
- Convert between float types
- Convert between array or slice
- Convert between two map
- Convert between time to string or integer types (int32, int64, int)
- Convert between struct types (same field name, or custom)
- Custom rule
- Global custom rule

## Usage

Convert directly:

```go
var src = []int32{1, 2, 3}
var dst []int
xconv.Convert(src, &dst)
```

Convert time to integer:

```go
var src = time.Now()
var dst int64
xconv.Convert(src, &dst)
```

Convert time to string by custom display (default display is "2006-01-02 15:04:05"):

```go
display := "2006-01-02 15"
var src = time.Now()
var dst string
xconv.NewConvertor(src).TimeFormat(display).Apply(&dst)
```

Convert struct to another (It will convert same field name, include anonymous 
fields. Fields must be exportable):

```go
type SrcStruct{
    A int32
}

type DstStruct{
    A int
}

var src = SrcStruct{A: 1}
var dst DstStruct
xconv.Convert(src, &dst)
```

Custom converting rule:

```go
type SrcStruct{
    A int
}

type DstStruct{
    A string
}

var src = SrcStruct{A: 1}
var dst DstStruct
NewConvertor(src).
    Rule(xconv.IntTypes, reflect.String,
        func(c *xconv.Convertor, src, dst reflect.Value) {
            dst.SetString(strconv.Itoa(int(src.Int())))
        }).
    Apply(&dst)
```

Or custom field rule:

```go
type SrcStruct{
    A int
}

type DstStruct{
    B int64
}

var src = SrcStruct{A: 1}
var dst DstStruct
NewConvertor(src).
    Field("B",
        func(srcObj SrcStruct) int64 {
            return int64(srcObj.A)
        })
    Apply(&dst)
```

Custom global rule:

```go
import (
    "reflect"
    "strconv"

    "github.com/howcrazy/xconv"
)

func init(){
    xconv.ConvertMap.
        TimeFormat("2006-01-02 15:04").
        Set(xconv.IntTypes, xconv.StringTypes,
            func(c *Convertor, src, dst reflect.Value) {
                dst.SetString(strconv.Itoa(int(src.Int())))
            })
}
```
