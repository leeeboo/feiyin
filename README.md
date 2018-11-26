#飞印打印机Golang SDK

飞印打印机：http://www.feyin.net/
相关API文档：https://www.showdoc.cc/feyin?page_id=350026008697847

使用方法：

```go
package main

import "fmt"
import "github.com/leeeboo/feiyin"


func main() {
	c, err := feiyin.NewClient(MEMBER_CODE, APPID, APP_SECRET, REDIS_ADDR) //redis用来做accesstoken的缓存，例如：127.0.0.1:6379
	if err != nil {
		panic(err)
	}

    //枚举已授权的商户清单
	data, err := c.Members()
	fmt.Println(data, err)

    //查询已授权商户信息
    data, err = c.Member(UID)
    fmt.Println(data, err)

    //解除绑定打印机
    err = c.DeviceUnbind(DEVICE_NO)
    fmt.Println(err)

    //绑定打印机
    err = c.DeviceBind(DEVICE_NO)
    fmt.Println(err)

    //发送打印消息
    //普通打印
    MsgNo, err := c.Print(DEVICE_NO, "测试测试测试测\nssssss试测试测试", "", "")
    fmt.Println(MsgNo, err)

    //模版打印

    tData := map[string]interface{}{
		"text1":  "第一个",
	}

    MsgNo, err = c.Print(DEVICE_NO, "", TEMPLATE_ID, tData)
    fmt.Println(MsgNo, err)

    //查看指定打印机状态
    data, err = c.Device(DEVICE_NO)
    fmt.Println(data, err)

    //查看所有打印机状态
    data, err = c.Devices()
    fmt.Println(data, err)

    //查询消息打印状态
    data, err = c.MsgStatus(MSG_NO)
    fmt.Println(data, err)

    //撤销未打印信息
    err = c.MsgCancel(MSG_NO)
    fmt.Println(err)

    //清除所有未打印消息
    count, err = c.DeviceClear(DEVICE_NO)
    fmt.Println(count, err)

    //创建标签打印模板
    content := `SIZE 60 mm,40 mm
			CLS
			TEXT 50,50,"4",0,1,1,"DEMO FOR TEXT"
			PRINT 1`

	id, err := c.TemplateAdd("test", content, "tsc", "test-desc")
	fmt.Println(id, err)

    //编辑标签打印模板
    content := `SIZE 60 mm,40 mm
			CLS
			TEXT 50,50,"4",0,1,1,"DEMO FOR TEXT {{name}}"
			PRINT 1`

	id, err := c.TemplateEdit(TEMPLATE_ID, "test", content, "tsc", "test-desc")
	fmt.Println(id, err)

    //获取指定的打印模板
    data, err = c.Template(TEMPLATE_ID)
    fmt.Println(data, err)

    //获取所有打印模板列表
    data, err = c.Templates()
    fmt.Println(data, err)
}
```
