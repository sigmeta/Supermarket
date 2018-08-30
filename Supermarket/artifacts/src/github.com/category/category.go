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

var logger = shim.NewLogger("Category")

// 每类商品 category struct
type Record struct {
	ID        string `json:ID`   // ID
	Name      string `json:Name` // full name
	StoreID   string `json:StoreID`
	StoreName string `json:StoreName`
	BarCode   string `json:BarCode`   //
	MeaUnit   string `json:MeaUnit`   // MeasurementUnit
	UnitPrice string `json:UnitPrice` // unit-price
	ShelfLife string `json:ShelfLife` // Quality guarantee period; shelf-life
	Stock	  string `json:Stock`     //Stock remains
	//Supplier  string		 	`json:Supplier`
	//Place     string        	`json:Place`     	// place of production
	CreateTime string        `json:CreateTime` // 创建时间
	History    []HistoryItem `json:History`    //
}

// 历史item结构
type HistoryItem struct {
	TxId   string `json:"txId"`
	Record Record `json:"record"`
}

// 前缀
const Record_Prefix = "Cate_"
const Stock_Prefix = "Stock_"

// composite keys
const IndexName = "storeID~CateID"
const CommIndexName = "storeID~CommID"

// chaincode response结构
type chaincodeRet struct {
	Code int    // 0 success otherwise 1
	Des  string //description
}

// 根据ID取出记录
func (a *CategoryChaincode) getRecord(stub shim.ChaincodeStubInterface, key string) (Record, bool) {
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
func (a *CategoryChaincode) putRecord(stub shim.ChaincodeStubInterface, key string, record Record) ([]byte, bool) {

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

// CategoryChaincode example Category Chaincode implementation
type CategoryChaincode struct {
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

func (t *CategoryChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Category Chaincode Init ###########")
	//val, ok, err := cid.GetAttributeValue(stub, "type")
	//logger.Info(val,ok,err)
	//res := getRetByte(0, "############"+string(val)+string(err.Error()))
	//return shim.Success(res)
	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *CategoryChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### chaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	logger.Info("%s%s", "CategoryChaincode function=", function)
	logger.Info("%s%s", "CategoryChaincode args=", args)
	if function == "insert" {
		// 插入信息
		return t.insert(stub, args)
	} else if function == "queryByID" {
		// 根据编号查询
		return t.queryByID(stub, args)
	} else if function == "query" {
		// 根据完整key查询
		return t.query(stub, args)
	} else if function == "change" {
		// 更改信息
		return t.change(stub, args)
	} else if function == "delete" {
		// 删除记录
		return t.delete(stub, args)
	} else if function == "insertStock" {
		// 初始化插入库存
		return t.insertStock(stub, args)
	} else if function == "changeStock" {
		// 修改库存
		return t.changeStock(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument. Wrong action: %v", args[0])
	res := getRetString(1, "Unknown action")
	return shim.Error(res)
}

// 加入新记录
// args: 0 - {Record Object}
func (a *CategoryChaincode) insert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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

	CateStoreKey, err := stub.CreateCompositeKey(IndexName, []string{Record_Prefix + record.ID, record.StoreID})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insert: CreateCompositeKey failed")
		return shim.Error(res)
	}
	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, CateStoreKey)
	if existbl {
		res := getRetString(1, "Chaincode Invoke insert failed : the recordNo has exist ")
		return shim.Error(res)
	}
	//13位时间戳
	record.CreateTime = strconv.FormatInt(time.Now().Unix(), 10)
	record.Stock = "0"
	// 保存记录
	_, bl := a.putRecord(stub, CateStoreKey, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke insert put record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke insert success")
	return shim.Success(res)
}

// 根据商品category ID查找记录
//  0 - Record_No ;
func (a *CategoryChaincode) queryByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "CategoryChaincode queryByRecordNo args!=1")
		return shim.Error(res)
	}
	// 取得所有索引
	recordsIterator, err := stub.GetStateByPartialCompositeKey(IndexName, []string{Record_Prefix + args[0]})
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

// 根据组合key查找记录
//  0 - Category ID, 1 - Store ID ;
func (a *CategoryChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		res := getRetString(1, "CategoryChaincode queryByRecordNo args!=1")
		return shim.Error(res)
	}
	CateStoreKey, err := stub.CreateCompositeKey(IndexName, []string{Record_Prefix + args[0], args[1]})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insert: CreateCompositeKey failed")
		return shim.Error(res)
	}
	// 取得该记录
	record, bl := a.getRecord(stub, CateStoreKey)
	if !bl {
		res := getRetString(1, "CategoryChaincode queryByRecordNo get record error")
		return shim.Error(res)
	}

	// 取得历史: 通过fabric api取得该商品的变更历史
	resultsIterator, err := stub.GetHistoryForKey(CateStoreKey)
	if err != nil {
		res := getRetString(1, "CategoryChaincode queryByRecordNo GetHistoryForKey error")
		return shim.Error(res)
	}
	defer resultsIterator.Close()

	var history []HistoryItem
	var hisRecord Record
	for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
			res := getRetString(1, "CategoryChaincode queryByRecordNo resultsIterator.Next() error")
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
	// 将历史做为记录的一个属性 一同返回
	record.History = history

	b, err := json.Marshal(record)
	if err != nil {
		res := getRetString(1, "CategoryChaincode Marshal queryByRecordNo recordList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// 修改记录
// args: 0 - {Record Object}
func (a *CategoryChaincode) change(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	CateStoreKey, err := stub.CreateCompositeKey(IndexName, []string{Record_Prefix + record.ID, record.StoreID})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insert: CreateCompositeKey failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	_, existbl := a.getRecord(stub, CateStoreKey)
	if !existbl {
		res := getRetString(1, "Chaincode Invoke change failed : change without existed record")
		return shim.Error(res)
	}

	// 保存记录
	_, bl := a.putRecord(stub, CateStoreKey, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke change put record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke change success")
	return shim.Success(res)
}

// 删除记录
// args: 0 - ID, 1 - Store ID
func (a *CategoryChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		res := getRetString(1, "Chaincode Invoke delete args!=1")
		return shim.Error(res)
	}

	CateStoreKey, err := stub.CreateCompositeKey(IndexName, []string{Record_Prefix + args[0], args[1]})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insert: CreateCompositeKey failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	record, existbl := a.getRecord(stub, CateStoreKey)
	if !existbl {
		res := getRetString(1, "Chaincode Invoke delete failed : delete without existed record ")
		return shim.Error(res)
	}

	err = stub.DelState(CateStoreKey)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke delete delete record failed")
		return shim.Error(res)
	}

	b, err := json.Marshal(record)
	if err != nil {
		res := getRetString(1, "CategoryChaincode Marshal delete recordList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// insert stock
// args: 0 - ID, 1 - Store ID, 2 - quantity
func (a *CategoryChaincode) insertStock(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		res := getRetString(1, "Chaincode Invoke insertStock args!=1")
		return shim.Error(res)
	}

	CateStoreKey, err := stub.CreateCompositeKey(IndexName, []string{Stock_Prefix + args[0], args[1]})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insertStock: CreateCompositeKey failed")
		return shim.Error(res)
	}

	//查找是否key已存在
	_, err = stub.GetState(CateStoreKey)
	if err == nil {
		res := getRetString(1, "Chaincode Invoke insertStock failed : stock has existed ")
		return shim.Error(res)
	}

	err = stub.PutState(CateStoreKey, []byte(args[2]))
	if err != nil {
		res := getRetString(1, "Chaincode Invoke insertStock failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke insertStock success")
	return shim.Success(res)
}

// change stock
// args: 0 - Category ID, 1 - Store ID, 2 - quantity, 3 - "add" or "reduce"
func (a *CategoryChaincode) changeStock(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		res := getRetString(1, "Chaincode Invoke changeStock args!=1")
		return shim.Error(res)
	}

	CateStoreKey, err := stub.CreateCompositeKey(IndexName, []string{Record_Prefix + args[0], args[1]})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke changeStock: CreateCompositeKey failed")
		return shim.Error(res)
	}

	//根据ID 查找是否ID已存在
	record, existbl := a.getRecord(stub, CateStoreKey)
	if !existbl {
		res := getRetString(1, "Chaincode Invoke changeStock failed : change without existed record")
		return shim.Error(res)
	}

	stockFloat, err := strconv.ParseFloat(string(record.Stock), 32)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke changeStock failed : cannot parse stock to float")
		return shim.Error(res)
	}
	quantityFloat, err := strconv.ParseFloat(args[2], 32)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke changeStock failed : cannot parse quantity from string to float")
		return shim.Error(res)
	}
	if args[3] == "add" {
		stockFloat = stockFloat + quantityFloat
	} else if args[3] == "reduce" {
		stockFloat = stockFloat - quantityFloat
	} else {
		res := getRetString(1, "Chaincode Invoke changeStock failed : args[3] should be add or reduce")
		return shim.Error(res)
	}

	record.Stock = strconv.FormatFloat(stockFloat, 'f', -1, 32)

	// 保存记录
	_, bl := a.putRecord(stub, CateStoreKey, record)
	if !bl {
		res := getRetString(1, "Chaincode Invoke changeStock put record failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke changeStock success")
	return shim.Success(res)
}

func main() {
	err := shim.Start(new(CategoryChaincode))
	if err != nil {
		logger.Errorf("Error starting Category chaincode: %s", err)
	}
}
