package main

import (
	"bytes"
	"log"
	"time"
    "encoding/json"
    "fmt"
	"github.com/streadway/amqp"
    "strconv"
    "net/url"
    "net/http"
    "crypto/md5"
    "encoding/hex"
    "io/ioutil"
    "math/rand"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
const APPKEY = "5b433b1f92d41bba340a5bb47464ce32" //您申请的APPKEY
func main() {
	conn, err := amqp.Dial("amqp://guest:guest@3.81.214.206:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"pay_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dot_count := bytes.Count(d.Body, []byte("."))
			/******************解析json*****************/
			    var mapdata map[string]interface{}
			    // 将字符串反解析为字典
			    json.Unmarshal([]byte(d.Body), &mapdata)
			    orderid := mapdata["orderid"].(string)
			    totalprice := mapdata["totalprice"].(string)
			    code := mapdata["code"].(string)
			    fmt.Println("orderid:",mapdata["orderid"])
			    fmt.Println("totalprice:",mapdata["totalprice"])
			    fmt.Println("code:",mapdata["code"])
            /******************解析json*****************/

            /******************查询订单状态确定是否为2*****************/
                orderidstatus := getorderStatus(orderid)
                 if orderidstatus != "2" {
                     fmt.Print("订单状态2可以减币操作:",orderidstatus)
                                /******************发送订单*****************/
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
                                isSuccess :=subNum(token,resultNum)
                                fmt.Println(isSuccess)
                                if isSuccess == true {
                                    fmt.Println("减币成功")
                                    /************发送请求给聚合************/
                                    data,err:=Post(juheURL,param)
                                    if err!=nil{
                                        fmt.Errorf("请求聚合失败,错误信息:\r\n%v",err)
                                    }else{
                                        var netReturn map[string]interface{}
                                        json.Unmarshal(data,&netReturn)
                                        fmt.Printf("请求聚合返回结果:\r\n%v",netReturn)
                                    }
                                    /***********发送请求给聚合************/
                                }else{
                                    fmt.Println("减币失败")   
                                }
                                /***************减币成功才发送订单*****************/             
                               /******************发送订单*****************/
                     
                 }else{
                     fmt.Print("订单状态不对:",orderidstatus)
                 }
            /******************查询订单状态确定是否为2*****************/


			t := time.Duration(dot_count)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}



//查询订单状态，确定状态是否是2
func getorderStatus(orderid string)(s string) {
    //请求地址
    juheURL :="http://op.juhe.cn/trainTickets/orderStatus"
    
    //初始化参数
    param:=url.Values{}
    
    //配置请求参数,方法内部已处理urlencode问题,中文参数可以直接传参
    param.Set("dtype","json") //返回的格式，json或xml，默认json
    param.Set("key",APPKEY)
    param.Set("orderid",orderid) //发车日期，如：2015-07-01（务必按照此格式）
    
    var status string
    //发送请求
    data,err:=Post(juheURL,param)
    if err!=nil{
        fmt.Errorf("请求失败,错误信息:\r\n%v",err)

    }else{
        var netReturn map[string]interface{}
        json.Unmarshal(data,&netReturn)
        status = netReturn["result"].(map[string]interface{})["status"].(string)
        
    }

    return status
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

//用户减币函数
func subNum(taken string,num float64)(y bool) {
    data := make(map[string]interface{})
    data["access"] = taken
    data["mark"] = "TGV"
    data["num"] = num
    data["order_no"] = GetRandomString(6)
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
     
    var isSuccess bool
    
    isSuccess = netReturn["isSuccess"].(bool)
    return isSuccess
}



// 生成32位MD5
func MD5(text string) string {
    ctx := md5.New()
    ctx.Write([]byte(text))
    return hex.EncodeToString(ctx.Sum(nil))
}

//获取当前时间
func GetTime() string {
    const shortForm = "20060102150405"
    t := time.Now()
    temp := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
    str := temp.Format(shortForm)
    return str
}
// 随机生成置顶位数的大写字母和数字的组合
func  GetRandomString(l int) string {
    //str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    str := "0123456789"
    bytes := []byte(str)
    result := []byte{}
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    for i := 0; i < l; i++ {
        result = append(result, bytes[r.Intn(len(bytes))])
    }
    return GetTime() + string(result)
}

// get 网络请求
func Get(apiURL string,params url.Values)(rs[]byte ,err error){
    var Url *url.URL
    Url,err=url.Parse(apiURL)
    if err!=nil{
        fmt.Printf("解析url错误:\r\n%v",err)
        return nil,err
    }
    //如果参数中有中文参数,这个方法会进行URLEncode
    Url.RawQuery=params.Encode()
    resp,err:=http.Get(Url.String())
    if err!=nil{
        fmt.Println("err:",err)
        return nil,err
    }
    defer resp.Body.Close()
    return ioutil.ReadAll(resp.Body)
}
 
// post 网络请求 ,params 是url.Values类型
func Post(apiURL string, params url.Values)(rs[]byte,err error){
    resp,err:=http.PostForm(apiURL, params)
    if err!=nil{
        return nil ,err
    }
    defer resp.Body.Close()
    return ioutil.ReadAll(resp.Body)
}