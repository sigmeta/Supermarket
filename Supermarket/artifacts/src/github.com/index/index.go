/*
全国索引
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

var logger = shim.NewLogger("Index")

// 索引
type Record struct {
	ID         string `json:ID`         //ID
	Channel    string `json:Channel`    //
	Chaincode  string `json:Chaincode`  //
	CreateTime string `json:CreateTime` //创建时间
}

// search表的映射名
const IdChannelChaincodeKeyStruct = "ID~Channel~Chaincode"

// chaincode response结构
type chaincodeRet struct {
	Code int    // 0 success otherwise 1
	Des  string //description
}

// 根据票号取出票据
func (a *IndexChaincode) getRecord(stub shim.ChaincodeStubInterface, record_No string) (Record, bool) {
	var record Record
	key := record_No
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

// 保存票据
func (a *IndexChaincode) putRecord(stub shim.ChaincodeStubInterface, key string, record Record) ([]byte, bool) {

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

// IndexChaincode example Index Chaincode implementation
type IndexChaincode struct {
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

func (t *IndexChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### GovernmentAffairs Init ###########")
	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *IndexChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### chaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	logger.Info("%s%s", "IndexChaincode function=", function)
	logger.Info("%s%s", "IndexChaincode args=", args)
	val, ok, err := cid.GetAttributeValue(stub, "type")
	logger.Info(val, ok, err)
	if function == "insert" {
		// 发布提案
		return t.insert(stub, args)
	} else if function == "queryByID" {
		// 根据编号查询
		return t.queryByID(stub, args)
	} else if function == "delete" {
		// 根据编号查询
		return t.delete(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument. Wrong action: %v", args[0])
	res := getRetString(1, "Unknown action: "+args[0])
	return shim.Error(res)
}

// 加入新记录
// args: 0 - {Record Object}
func (a *IndexChaincode) insert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	IdChannelChaincodeKey, err := stub.CreateCompositeKey(IdChannelChaincodeKeyStruct, []string{record.ID, record.Channel, record.Chaincode})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insert: CreateCompositeKey failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, IdChannelChaincodeKey)
	if existbl {
		res := getRetString(1, "Chaincode Invoke insert failed : the index has exist ")
		return shim.Error(res)
	}
	record.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)

	// 保存记录
	_, bl := a.putRecord(stub, IdChannelChaincodeKey, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke insert put record failed")
		return shim.Error(res)
	}
	val, ok, err := cid.GetAttributeValue(stub, "type")
	logger.Info(val, ok, err)
	res := getRetByte(0, "invoke insert success "+string(val))
	return shim.Success(res)
}

// 根据ID查找记录
//  0 - Record_No ;
func (a *IndexChaincode) queryByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "IndexChaincode queryByRecordNo args!=1")
		return shim.Error(res)
	}
	// 取得所有索引
	recordsIterator, err := stub.GetStateByPartialCompositeKey(IdChannelChaincodeKeyStruct, []string{args[0]})
	if err != nil {
		res := getRetString(1, "IndexChaincode queryByID get record error")
		return shim.Error(res)
	}
	defer recordsIterator.Close()
	// 索引列表
	var recordList = []Record{}
	for recordsIterator.HasNext() {
		var record Record
		kv, _ := recordsIterator.Next()
		// kv.Value为内容
		err = json.Unmarshal(kv.Value, &record)
		if err != nil {
			res := getRetString(1, "WorkflowChaincode queryByID unmarshal failed")
			return shim.Error(res)
		}
		// 取得ID与查询ID相同的加入列表
		if record.ID == args[0] {
			recordList = append(recordList, record)
		}
	}

	b, err := json.Marshal(recordList)
	if err != nil {
		res := getRetString(1, "IndexChaincode Marshal queryByRecordNo recordList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// 删除记录
// args: 0 - {Record Object}
func (a *IndexChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "Chaincode Invoke insert args!=1")
		return shim.Error(res)
	}

	var record Record
	err := json.Unmarshal([]byte(args[0]), &record)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete unmarshal failed")
		return shim.Error(res)
	}
	IdChannelChaincodeKey, err := stub.CreateCompositeKey(IdChannelChaincodeKeyStruct, []string{record.ID, record.Channel, record.Chaincode})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete: CreateCompositeKey failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, IdChannelChaincodeKey)
	if !existbl {
		res := getRetString(1, "Chaincode Invoke delete failed : delete an nonexistent record ")
		return shim.Error(res)
	}

	// 保存记录
	err = stub.DelState(IdChannelChaincodeKey)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke delete success")
	return shim.Success(res)
}

func main() {
	err := shim.Start(new(IndexChaincode))
	if err != nil {
		logger.Errorf("Error starting Index chaincode: %s", err)
	}
}
