package main

import (
	"github.com/gin-gonic/gin" 
    "fmt"
    "trainTickets/utils"
    "net/url"
    "encoding/json"
    "bytes"
    "io/ioutil"
    "net/http"
    "strconv"
)
//rabbitmq使用的错误输出
func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
    }
}

const APPKEY = "5b433b1f92d41bba340a5bb47464ce32" //您申请的APPKEY
//站点简码查询
func main() {
	//获取参数
	orderid := ctx.PostForm("orderid")
    totalprice := ctx.PostForm("totalprice")
    code := ctx.PostForm("code")
    token := getAccess(code)//根据前端传来的code获取token
	lastprice :=  getLastprice()//请求接口获取最新价格
    /**************将价格转成浮点数**************/
    total, _ := strconv.ParseFloat(totalprice, 64)
    last, _ := strconv.ParseFloat(lastprice, 64)
    resultdata := total/last
    resultstr := fmt.Sprintf("%.4f",resultdata)
    resultNum, _:= strconv.ParseFloat(resultstr, 64)
    fmt.Println("resultNum:",resultNum)
    /**************将价格转成浮点数**************/
	//请求地址
	juheURL :="http://op.juhe.cn/trainTickets/pay"
	
	//初始化参数
	param:=url.Values{}
	
	//配置请求参数,方法内部已处理urlencode问题,中文参数可以直接传参
	param.Set("dtype","json") //返回的格式，json或xml，默认json
	param.Set("key",APPKEY)
	param.Set("orderid",orderid) //订单号，提交订单时会返回
	
    /***************先减币成功才发送订单*****************/
    data := make(map[string]interface{})
    data["access"] = token
    data["mark"] = "TGV"
    data["num"] = resultNum
    data["order_no"] = utils.GetRandomString(6)
    bytesData, err := json.Marshal(data)
    if err != nil {
        fmt.Println(err.Error() )
        return
    }
    reader := bytes.NewReader(bytesData)
    url := "http://47.52.47.73:8112/auth/auth/subNum"
    request, err := http.NewRequest("POST", url, reader)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    request.Header.Set("Content-Type", "application/json;charset=UTF-8")
    client := http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    //返回结果
    var netReturn map[string]interface{}
    json.Unmarshal(respBytes,&netReturn)
    if netReturn["isSuccess"]==true{
        fmt.Printf("减币成功:\r\n%v",netReturn)
        /*****************减币成功********************/
        
        /************发送请求给聚合************/
		data,err:=utils.Post(juheURL,param)
		if err!=nil{
			fmt.Errorf("请求失败,错误信息:\r\n%v",err)
			ctx.JSON(404, gin.H{
				"error_code": "404",
				"message": err,
			})
		}else{
			var netReturn map[string]interface{}
			json.Unmarshal(data,&netReturn)

			ctx.JSON(200, gin.H{
				"error_code": netReturn["error_code"],
				"message": netReturn["reason"],
				"result":netReturn["result"],
			})
		}
		/************************************/

    }else{
    	/*****************减币失败********************/
            ctx.JSON(404, gin.H{
			"isSuccess": netReturn["isSuccess"],
			"message": netReturn["message"],
		})
    }
    /***************减币成功才发送订单*****************/

}








//获取最新汇率
func getLastprice()(lastprice string) {
    huili := make(map[string]interface{})
    huili["currency_mark"] = "USDT"
    huili["currency_trade_mark"] = "TGV"
    bytesData, err := json.Marshal(huili)
    if err != nil {
        fmt.Println(err.Error() )
        return
    }
    reader := bytes.NewReader(bytesData)
    url := "https://m.51tg.vip/api/currency/currency/getLastData"
    request, err := http.NewRequest("POST", url, reader)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    request.Header.Set("Content-Type", "application/json;charset=UTF-8")
    client := http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    var data string
    //返回结果
    var netReturn map[string]interface{}
    json.Unmarshal(respBytes,&netReturn)
    if netReturn["isSuccess"]==true{
        last_price := netReturn["data"].(map[string]interface{})["lastData"].(map[string]interface{})["last_price"]
        // fmt.Println("获取最新价是:\r\n",last_price)
        data = last_price.(string)
    }else{
        fmt.Println("获取最新价失败")
    }
    return data
}
//根据前端传来的code获取token
func getAccess(code string)(token string) {
    data := make(map[string]interface{})
    data["code"] = code
    bytesData, err := json.Marshal(data)
    if err != nil {
        fmt.Println(err.Error() )
        return
    }
    reader := bytes.NewReader(bytesData)
    url := "http://47.52.47.73:8112/auth/auth/getAccess"
    request, err := http.NewRequest("POST", url, reader)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    request.Header.Set("Content-Type", "application/json;charset=UTF-8")
    client := http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    respBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    var tokendata string
    //返回结果
    var netReturn map[string]interface{}
    json.Unmarshal(respBytes,&netReturn)
    if netReturn["isSuccess"]==true{
        fmt.Printf("获取token:\r\n%v",netReturn["data"].(map[string]interface{})["auth"].(map[string]interface{})["access"])
        access := netReturn["data"].(map[string]interface{})["auth"].(map[string]interface{})["access"]
        tokendata = access.(string)
    }else{
        fmt.Println("获取token失败")
    }
    
    return tokendata
}