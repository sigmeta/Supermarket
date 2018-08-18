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

var logger = shim.NewLogger("Commodity")

// 每个商品 commodity struct
type Record struct {
	ID         string `json:ID`        // ID
	Name       string `json:Name`      // full name
	Commodity  string `json:Commodity` // category id
	StoreID    string `json:StoreID`
	StoreName  string `json:StoreName`
	Supplier   string `json:Supplier`
	Place      string `json:Place`      // place of production
	Date       string `json:Date`       //date of production
	CreateTime string `json:CreateTime` // 创建时间
}

// 历史item结构
type HistoryItem struct {
	TxId   string `json:"txId"`
	Record Record `json:"record"`
}

// 前缀
const Record_Prefix = "Comm_"

// composite keys
const IndexName = "storeID~CommID"

// chaincode response结构
type chaincodeRet struct {
	Code int    // 0 success otherwise 1
	Des  string //description
}

// 根据ID取出记录
func (a *CommodityChaincode) getRecord(stub shim.ChaincodeStubInterface, key string) (Record, bool) {
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
func (a *CommodityChaincode) putRecord(stub shim.ChaincodeStubInterface, key string, record Record) ([]byte, bool) {

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

// CommodityChaincode example Commodity Chaincode implementation
type CommodityChaincode struct {
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

func (t *CommodityChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Commodity Chaincode Init ###########")
	//val, ok, err := cid.GetAttributeValue(stub, "type")
	//logger.Info(val,ok,err)
	//res := getRetByte(0, "############"+string(val)+string(err.Error()))
	//return shim.Success(res)
	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *CommodityChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### chaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	logger.Info("%s%s", "CommodityChaincode function=", function)
	logger.Info("%s%s", "CommodityChaincode args=", args)
	if function == "insert" {
		// 插入信息
		return t.insert(stub, args)
	} else if function == "queryByID" {
		// 根据编号查询
		return t.queryByID(stub, args)
	} else if function == "delete" {
		// 删除记录
		return t.delete(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument. Wrong action: %v", args[0])
	res := getRetString(1, "Unknown action")
	return shim.Error(res)
}

// 加入新记录
// args: 0 - {Record Object}
func (a *CommodityChaincode) insert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

	res := getRetByte(0, "invoke insert success")
	return shim.Success(res)
}

// 根据ID查找记录
//  0 - Commodity ID
func (a *CommodityChaincode) queryByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "CommodityChaincode queryByRecordNo args!=1")
		return shim.Error(res)
	}

	// 取得该记录
	record, bl := a.getRecord(stub, Record_Prefix+args[0])
	if !bl {
		res := getRetString(1, "CommodityChaincode queryByRecordNo get record error")
		return shim.Error(res)
	}

	b, err := json.Marshal(record)
	if err != nil {
		res := getRetString(1, "CommodityChaincode Marshal queryByRecordNo recordList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// TODO unused
// 修改记录
// args: 0 - {Record Object}
func (a *CommodityChaincode) change(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "Chaincode Invoke change args!=1")
		return shim.Error(res)
	}

	var record Record
	err := json.Unmarshal([]byte(args[0]), &record)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke change unmarshal failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, Record_Prefix+record.ID)
	if !existbl {
		res := getRetString(1, "Chaincode Invoke change failed : change without existed record")
		return shim.Error(res)
	}

	// 保存记录
	_, bl := a.putRecord(stub, Record_Prefix+record.ID, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke change put record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke change success")
	return shim.Success(res)
}

// 删除记录
// args: 0 - ID
func (a *CommodityChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	err := shim.Start(new(CommodityChaincode))
	if err != nil {
		logger.Errorf("Error starting Commodity chaincode: %s", err)
	}
}
