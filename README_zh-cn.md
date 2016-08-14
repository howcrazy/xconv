# XConv

一个Golang的类型转换工具，支持在各种不同的类型间进行转换（如基本类型、结构体、数组或列表、
map等），并支持自定义转换。

## 特性

- 整数类型间转换
- 浮点类型转换
- 数组或列表转换
- map转换
- 将时间转换成整数（unix）或字符串，也支持反向转换
- 将一个结构体类型换换成另一个类型（复制同名属性）
- 自定义转换规则
- 自定义全局转换规则

## 示例

按照默认规则直接转换：

```go
var src = []int32{1, 2, 3}
var dst []int
xconv.Convert(src, &dst)
```

将时间转换成整数：

```go
var src = time.Now()
var dst int64
xconv.Convert(src, &dst)
```

将时间转换成自定义格式的字符串（默认格式为"2006-01-02 15:04:05"）：

```go
display := "2006-01-02 15"
var src = time.Now()
var dst string
xconv.NewConvertor(src).TimeFormat(display).Apply(&dst)
```

结构体转换（自动转换同名属性，包括匿名属性，必须是可导出的）：

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

自定义转换规则：

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

或修改指定属性的转换规则：

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

自定义全局转换规则：

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
