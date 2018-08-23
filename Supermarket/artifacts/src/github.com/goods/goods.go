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

var logger = shim.NewLogger("Goods")

// 前缀
const Cate_Prefix = "Cate_"
const Comm_Prefix = "Comm_"

// composite keys
const CateIndexName = "storeID~CateID"
const CommIndexName = "storeID~CommID"

// chaincode response结构
type chaincodeRet struct {
	Code int    // 0 success otherwise 1
	Des  string //description
}


// GoodsChaincode example Goods Chaincode implementation
type GoodsChaincode struct {
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

func (t *GoodsChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Goods Chaincode Init ###########")
	//val, ok, err := cid.GetAttributeValue(stub, "type")
	//logger.Info(val,ok,err)
	//res := getRetByte(0, "############"+string(val)+string(err.Error()))
	//return shim.Success(res)
	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *GoodsChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### chaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	logger.Info("%s%s", "GoodsChaincode function=", function)
	logger.Info("%s%s", "GoodsChaincode args=", args)
	if function == "purchase" {
		// 插入信息
		return t.purchase(stub, args)
	} else if function == "queryByID" {
		// 根据编号查询
		return t.queryByID(stub, args)
	} else if function == "change" {
		// 更改信息
		return t.change(stub, args)
	} else if function == "delete" {
		// 删除记录
		return t.delete(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument. Wrong action: %v", args[0])
	res := getRetString(1, "Unknown action")
	return shim.Error(res)
}

// 进货
// args: 0 - channel name, 1 - store id, 2- category id, 3- {commodity object}
func (a *GoodsChaincode) purchase(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {

		res := getRetString(1, "Chaincode Invoke insert args!=1")
		return shim.Error(res)
	}

	newArgs:=[][]byte{[]byte()}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, record.ID)
	if existbl {
		res := getRetString(1, "Chaincode Invoke insert failed : the recordNo has exist ")
		return shim.Error(res)
	}
	//13位时间戳
	record.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)

	// 保存记录
	_, bl := a.putRecord(stub, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke insert put record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke insert success")
	return shim.Success(res)
}

// 根据ID查找记录
//  0 - Record_No ;
func (a *GoodsChaincode) queryByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "GoodsChaincode queryByRecordNo args!=1")
		return shim.Error(res)
	}
	// 取得该票据
	record, bl := a.getRecord(stub, args[0])
	if !bl {
		res := getRetString(1, "GoodsChaincode queryByRecordNo get record error")
		return shim.Error(res)
	}

	// 取得背书历史: 通过fabric api取得该票据的变更历史
	resultsIterator, err := stub.GetHistoryForKey(Record_Prefix + args[0])
	if err != nil {
		res := getRetString(1, "GoodsChaincode queryByRecordNo GetHistoryForKey error")
		return shim.Error(res)
	}
	defer resultsIterator.Close()

	var history []HistoryItem
	var hisRecord Record
	for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
			res := getRetString(1, "GoodsChaincode queryByRecordNo resultsIterator.Next() error")
			return shim.Error(res)
		}

		var hisItem HistoryItem
		hisItem.TxId = historyData.TxId               //copy transaction id over
		json.Unmarshal(historyData.Value, &hisRecord) //un stringify it aka JSON.parse()
		if historyData.Value == nil {                 //record has been deleted
			var emptyRecord Record
			hisItem.Record = emptyRecord //copy nil marble
		} else {
			json.Unmarshal(historyData.Value, &hisRecord) //un stringify it aka JSON.parse()
			hisItem.Record = hisRecord                    //copy record over
		}
		history = append(history, hisItem) //add this tx to the list
	}
	// 将背书历史做为票据的一个属性 一同返回
	record.History = history

	b, err := json.Marshal(record)
	if err != nil {
		res := getRetString(1, "GoodsChaincode Marshal queryByRecordNo recordList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// 修改记录
// args: 0 - {Record Object}
func (a *GoodsChaincode) change(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	_, existbl := a.getRecord(stub, record.ID)
	if !existbl {
		res := getRetString(1, "Chaincode Invoke change failed : change without existed record")
		return shim.Error(res)
	}

	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		res := getRetString(1, "Chaincode Invoke change failed :get time stamp failed ")
		return shim.Error(res)
	}
	logger.Error("%s", timestamp)
	record.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)

	// 保存记录
	_, bl := a.putRecord(stub, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke change put record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke change success")
	return shim.Success(res)
}

// 加入新记录
// args: 0 - {Record Object}
func (a *GoodsChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "Chaincode Invoke delete args!=1")
		return shim.Error(res)
	}

	var record Record
	err := json.Unmarshal([]byte(args[0]), &record)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete unmarshal failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, record.ID)
	if !existbl {
		res := getRetString(1, "Chaincode Invoke delete failed : delete without existed record ")
		return shim.Error(res)
	}

	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete failed :get time stamp failed ")
		return shim.Error(res)
	}
	logger.Error("%s", timestamp)
	record.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)

	// 保存记录
	err = stub.DelState(Record_Prefix + record.ID)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete delete record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke delete success")
	return shim.Success(res)
}

func main() {
	err := shim.Start(new(GoodsChaincode))
	if err != nil {
		logger.Errorf("Error starting Goods chaincode: %s", err)
	}
}
