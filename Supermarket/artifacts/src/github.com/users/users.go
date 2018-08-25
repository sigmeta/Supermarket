/*
商品记录
每类商品有编号，用于记录价格、剩余量等信息
每个商品有单独编号，用于记录进货时间等信息
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

var logger = shim.NewLogger("Users")

// user
type Record struct {
	ID       string `json:ID`   // ID
	Name     string `json:Name` // full name
	Password string `json:Password`
	Coupon   string `json:Coupon` // category id
	//BlackList 	bool 		`json:StoreID`
	VIP        string `json:VIP`
	Phone      string `json:Phone`
	Cost       string `json:Cost`       // 消费总金额
	CreateTime string `json:CreateTime` // 创建时间
}

// VIP level
const (
	VIPLevel0 = "level0"
	VIPLevel1 = "level1"
	VIPLevel2 = "level2"
	VIPLevel3 = "level3"
	VIPLevel4 = "level4"
)

// 历史item结构
type HistoryItem struct {
	TxId   string `json:"txId"`
	Record Record `json:"record"`
}

// 前缀
const Record_Prefix = "User_"
const Password_Prefix = "Pwd_"

// composite keys
const IndexName = "storeID~CommID"

// chaincode response结构
type chaincodeRet struct {
	Code int    // 0 success otherwise 1
	Des  string //description
}

// 根据ID取出记录
func (a *UsersChaincode) getRecord(stub shim.ChaincodeStubInterface, key string) (Record, bool) {
	var record Record
	b, err := stub.GetState(key)
	if b == nil {
		return record, false
	}
	err = json.Unmarshal(b, &record)
	if err != nil {
		return record, false
	}
	return record, true
}

// 保存记录
func (a *UsersChaincode) putRecord(stub shim.ChaincodeStubInterface, key string, record Record) ([]byte, bool) {

	byte, err := json.Marshal(record)
	if err != nil {
		return nil, false
	}

	err = stub.PutState(key, byte)
	if err != nil {
		return nil, false
	}
	return byte, true
}

// UsersChaincode example Users Chaincode implementation
type UsersChaincode struct {
}

// response message format
func getRetByte(code int, des string) []byte {
	var r chaincodeRet
	r.Code = code
	r.Des = des

	b, err := json.Marshal(r)

	if err != nil {
		fmt.Println("marshal Ret failed")
		return nil
	}
	return b
}

// response message format
func getRetString(code int, des string) string {
	var r chaincodeRet
	r.Code = code
	r.Des = des

	b, err := json.Marshal(r)

	if err != nil {
		fmt.Println("marshal Ret failed")
		return ""
	}
	logger.Infof("%s", string(b[:]))
	return string(b[:])
}

func (a *UsersChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Users Chaincode Init ###########")
	//val, ok, err := cid.GetAttributeValue(stub, "type")
	//logger.Info(val,ok,err)
	//res := getRetByte(0, "############"+string(val)+string(err.Error()))
	//return shim.Success(res)
	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (a *UsersChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### chaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	logger.Info("%s%s", "UsersChaincode function=", function)
	logger.Info("%s%s", "UsersChaincode args=", args)
	if function == "insert" {
		// 插入信息
		return a.insert(stub, args)
	} else if function == "queryByID" {
		// 根据编号查询
		return a.queryByID(stub, args)
	} else if function == "change" {
		// 修改记录
		return a.change(stub, args)
	} else if function == "delete" {
		// 删除记录
		return a.delete(stub, args)
	} else if function == "login" {
		// 登录
		return a.login(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument. Wrong action: %v", args[0])
	res := getRetString(1, "Unknown action")
	return shim.Error(res)
}

// 加入新记录, register
// args: 0 - {Record Object}
func (a *UsersChaincode) insert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "Chaincode Invoke insert args!=1")
		return shim.Error(res)
	}

	var record Record
	err := json.Unmarshal([]byte(args[0]), &record)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insert unmarshal failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, Record_Prefix+record.ID)
	if existbl {
		res := getRetString(1, "Chaincode Invoke insert failed : the recordNo has exist ")
		return shim.Error(res)
	}
	//13位时间戳
	record.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)

	// 保存记录
	_, bl := a.putRecord(stub, Record_Prefix+record.ID, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke insert put record failed")
		return shim.Error(res)
	}

	err = stub.PutState(Password_Prefix+record.ID, []byte(record.Password))
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insert put password failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke insert success")
	return shim.Success(res)
}

// login
// args: 0 - ID, 1 - password (md5)
func (a *UsersChaincode) login(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		res := getRetString(1, "Chaincode Invoke login args!=2")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	password, err := stub.GetState(Password_Prefix + args[0])
	if err != nil {
		res := getRetString(1, "Chaincode Invoke login failed : the ID does not exist ")
		return shim.Error(res)
	}

	if string(password) == args[1] {
		res := getRetByte(0, "success")
		return shim.Success(res)
	} else {
		res := getRetByte(0, "failed")
		return shim.Success(res)
	}

}

// 根据ID查找记录
//  0 - Users ID
func (a *UsersChaincode) queryByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "UsersChaincode queryByRecordNo args!=1")
		return shim.Error(res)
	}

	// 取得该记录
	record, bl := a.getRecord(stub, Record_Prefix+args[0])
	if !bl {
		res := getRetString(1, "UsersChaincode queryByRecordNo get record error")
		return shim.Error(res)
	}

	b, err := json.Marshal(record)
	if err != nil {
		res := getRetString(1, "UsersChaincode Marshal queryByRecordNo recordList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// TODO unused
// 修改记录
// args: 0 - ID, 1 - json field, 2 - new value
func (a *UsersChaincode) change(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		res := getRetString(1, "Chaincode Invoke change args!=3")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	record, existbl := a.getRecord(stub, Record_Prefix+args[0])
	if !existbl {
		res := getRetString(1, "Chaincode Invoke change failed : change without existed record")
		return shim.Error(res)
	}
	if args[1] == "Coupon" {
		record.Coupon = args[2]
	} else if args[1] == "VIP" {
		record.VIP = args[2]
	} else if args[1] == "Phone" {
		record.Phone = args[2]
	} else if args[1] == "Cost" {
		cost, err := strconv.ParseFloat(record.Cost, 32)
		if err != nil {
			res := getRetString(1, "Chaincode Invoke change failed : cannot convert record.Cost to float")
			return shim.Error(res)
		}
		add, err := strconv.ParseFloat(args[2], 32)
		if err != nil {
			res := getRetString(1, "Chaincode Invoke change failed : cannot convert record.Cost to float")
			return shim.Error(res)
		}
		record.Cost = strconv.FormatFloat(cost+add, 'f', 2, 32)
	} else {
		res := getRetString(1, "wrong field: "+args[1])
		return shim.Error(res)
	}

	// 保存记录
	_, bl := a.putRecord(stub, Record_Prefix+args[0], record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke change put record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke change success")
	return shim.Success(res)
}

// 删除记录
// args: 0 - ID
func (a *UsersChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "Chaincode Invoke delete args!=1")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, Record_Prefix+args[0])
	if !existbl {
		res := getRetString(1, "Chaincode Invoke delete failed : delete without existed record ")
		return shim.Error(res)
	}

	err := stub.DelState(Record_Prefix + args[0])
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete delete record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke delete success")
	return shim.Success(res)
}

func main() {
	err := shim.Start(new(UsersChaincode))
	if err != nil {
		logger.Errorf("Error starting Users chaincode: %s", err)
	}
}
