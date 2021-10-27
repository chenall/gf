
这是一个基于gf原版衍生的修改版本，尽量保持和上游同步．

## Why?
   
日常使用过程中碰到了一些问题需要对源码进行一些改动处理，这时一般会提ISSUE或PR给上游，在上游接受这个改动期间我需要使用这些新的特性功能，来回切换会比较麻烦。

所以就自己建了一个新库，并通过 go.mod 的`replace`功能使用自己的库。 

## 修改的功能列表（部份已经并入上游）
* [支持带`BOM`的配置文件](https://github.com/gogf/gf/pull/825)。

  >这个主要是使用`Windows`系统的情况下会碰到，`Windows`自带的记事本会*自作聪明*的在`UTF8`编码的文件头添加一个`BOM`，导致程序启动失败，这个改动会自动过滤这个`BOM`。

* ~~gdb 支持使用 :var 占位符~~。

  >主要是为了方便编写`SQL`语句。相关的讨论在 [issue 811](https://github.com/gogf/gf/issues/811)

* gcron 的日期域支持 L,W 标志。  
  >对 gcron 进行功能扩充 [issue 836](https://github.com/gogf/gf/issues/836)  
 
  `L` 最后一天
  
  `W` 工作日
  
  组合:  
  `LW` 最后一个工作日  
  `nW` 距离指定日期n最近的工作日(不跨月)
  
  例子
  
  0 0 9 LW * ?  最个月最后一个工作日的9点整执行
  
  0 0 9 1W * ?  最个月初第一个工作日的9点整执行


* `gview` 模板文件查找规则修改
  
  修改前: 配置文件指定的位置放在`最后`
  
  修改后: `优先`从配置文件指定的位置开始查找

  
* `gcfg` 配置文件查找规则修改  
  修改前: `gres` 内置文件优先  
  修改后: 文件系统优先  


* `gconv.Struct` 针对 int 类型允许通过标签 `base` 指定使用 10 进制进行转换
  
  ```go
  type TestInt struct {
    I1 int `base:"10"`
    I2 int `base:"10"`
    I3 int `base:"10"`
    I4 int `base:"10"`
  }
  
  func Test_ConvDecimal(t *testing.T) {
    t1 := TestInt{}
    e := gconv.Struct(g.Map{"i1": "0777", "i2": "-011", "i3": " 006", "i4": "0x123"}, &t1)
    gtest.Assert(e, nil)
    gtest.Assert(t1.I1, 777)
    gtest.Assert(t1.I2, -11)
    gtest.Assert(t1.I3, 6)
    gtest.Assert(t1.I4, 0)
  }
  ```
* gconv 从 string 或 []byte 转换时不使用 `json Unmarshal` **21.07.15**  
  注：原版会尝试使用 `json Unmarshal` 导致丢失 gconv 的一些特性 [issue 1227](https://github.com/gogf/gf/issues/1227#issuecomment-828460451)


* 2021-10-27 gcfg 添加 FindConfig 函数 
  >判断 `name` 配置文件是否存在，如果存在返回配置，否则返回 `nil` 可以指定多个备用配置文件.  
  >FindConfig(name string, file ...string)  
  > 用于某些特定的场合，从多个配置文件中选取一个可用的。

