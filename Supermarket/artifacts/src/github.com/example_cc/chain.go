/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

var logger = shim.NewLogger("example_cc0")

const (
	BillInfo_State_WaitTeacherSign = "WaitTeacherSign"
	BillInfo_State_WaitSchoolSign  = "WaitSchoolSign"
	BillInfo_State_TeacherSigned   = "TeacherSigned"
	BillInfo_State_TeacherReject   = "TeacherReject"
	BillInfo_State_SchoolSigned    = "TeacherSigned"
	BillInfo_State_SchoolReject    = "SchoolReject"
	BillInfo_State_OverDue         = "OverDue"
)
const HolderIdDayTimeBillTypeBillNoIndexName = "holderId~dayTime-billType-billNo"

// 票据
type Bill struct {
	BillInfoID         string        `json:BillInfoID`         //票据号码
	BillInfoType       string        `json:BillInfoType`       //票据类型
	BillInfoIsseDate   string        `json:BillInfoIsseDate`   //票据出票日期
	BillInfoDueDate    string        `json:BillInfoDueDate`    //票据到期日期
	DrwrCmID           string        `json:DrwrCmID`           //出票人证件号码
	DrwrAcct           string        `json:DrwrAcct`           //出票人名称
	WaitEndorserCmID   string        `json:WaitEndorserCmID`   //待背书人证件号码
	WaitEndorserAcct   string        `json:WaitEndorserAcct`   //待背书人名称
	RejectEndorserCmID string        `json:RejectEndorserCmID` //拒绝背书人证件号码
	RejectEndorserAcct string        `json:RejectEndorserAcct` //拒绝背书人名称
	State              string        `json:State`              //票据状态
	History            []HistoryItem `json:History`            //背书历史
}

// 背书历史item结构
type HistoryItem struct {
	TxId string `json:"txId"`
	Bill Bill   `json:"bill"`
}

// 票据key的前缀
const Bill_Prefix = "Bill_"

//超时违约key的前缀
const OverDue_Prefix = "OD_"

// search表的映射名
const IndexName = "holderName~billNo"

// chaincode response结构
type chaincodeRet struct {
	Code int    // 0 success otherwise 1
	Des  string //description
}

// 根据票号取出票据
func (a *WorkflowChaincode) getBill(stub shim.ChaincodeStubInterface, bill_No string) (Bill, bool) {
	var bill Bill
	key := Bill_Prefix + bill_No
	b, err := stub.GetState(key)
	if b == nil {
		return bill, false
	}
	err = json.Unmarshal(b, &bill)
	if err != nil {
		return bill, false
	}
	return bill, true
}

// 保存票据
func (a *WorkflowChaincode) putBill(stub shim.ChaincodeStubInterface, bill Bill) ([]byte, bool) {

	byte, err := json.Marshal(bill)
	if err != nil {
		return nil, false
	}

	err = stub.PutState(Bill_Prefix+bill.BillInfoID, byte)
	if err != nil {
		return nil, false
	}
	return byte, true
}

// WorkflowChaincode example Workflow Chaincode implementation
type WorkflowChaincode struct {
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

func (t *WorkflowChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### GovernmentAffairs Init ###########")
	return shim.Success(nil)

}

// Transaction makes payment of X units from A to B
func (t *WorkflowChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### chaincode Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	logger.Info("%s%s", "WorkflowChaincode function=", function)
	logger.Info("%s%s", "WorkflowChaincode args=", args)
	if function == "issue" {
		// 发布提案
		return t.issue(stub, args)
	} else if function == "accept_teacher" {
		// 老师签字
		return t.accept_teacher(stub, args)
	} else if function == "accept_school" {
		// 学校签字
		return t.accept_school(stub, args)
	} else if function == "reject" {
		// 拒绝签字
		return t.reject(stub, args)
	} else if function == "queryMyBill" {
		// 查询我的提案
		return t.queryMyBill(stub, args)
	} else if function == "queryByBillNo" {
		// 根据编号查询
		return t.queryByBillNo(stub, args)
	} else if function == "queryMyWaitBill" {
		// 查询等待我签字的提案
		return t.queryMyWaitBill(stub, args)
	} else if function == "checkDue" {
		//检查是否逾期
		return t.checkDue(stub, args)
	}

	logger.Errorf("Unknown action, check the first argument. Wrong action: %v", args[0])
	res := getRetString(1, "Unknown action")
	return shim.Error(res)
}

// 票据发布
// args: 0 - {Bill Object}
func (a *WorkflowChaincode) issue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "Chaincode Invoke issue args!=1")
		return shim.Error(res)
	}

	var bill Bill
	err := json.Unmarshal([]byte(args[0]), &bill)
	if err != nil {
		res := getRetString(1, "Chaincode Invoke issue unmarshal failed")
		return shim.Error(res)
	}
	// TODO 根据票号 查找是否票号已存在
	// TODO 如stat中已有同号票据 返回error message
	_, existbl := a.getBill(stub, bill.BillInfoID)
	if existbl {
		res := getRetString(1, "Chaincode Invoke issue failed : the billNo has exist ")
		return shim.Error(res)
	}

	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		res := getRetString(1, "Chaincode Invoke issue failed :get time stamp failed ")
		return shim.Error(res)
	}
	logger.Error("%s", timestamp)
	bill.BillInfoIsseDate = strconv.FormatInt(time.Now().Unix(), 10)

	// 更改票据信息和状态并保存票据:票据状态设为待背书
	bill.State = BillInfo_State_WaitTeacherSign
	// 保存票据
	_, bl := a.putBill(stub, bill)
	if !bl {
		res := getRetString(1, "Chaincode Invoke issue put bill failed")
		return shim.Error(res)
	}
	// 以持票人ID和票号构造复合key 向search表中保存 value为空即可 以便持票人批量查询
	holderNameBillNoIndexKey, err := stub.CreateCompositeKey(IndexName, []string{bill.DrwrCmID, bill.BillInfoID})
	if err != nil {
		res := getRetString(1, "Chaincode Invoke issue put search table failed")
		return shim.Error(res)
	}
	stub.PutState(holderNameBillNoIndexKey, []byte{0x00})

	// 以待背书人ID和票号构造复合key 向search表中保存 value为空即可 以便待背书人批量查询
	holderNameBillNoIndexKey, err = stub.CreateCompositeKey(IndexName, []string{bill.WaitEndorserCmID, bill.BillInfoID})
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Invoke endorse put search table failed")
		return shim.Error(res)
	}
	stub.PutState(holderNameBillNoIndexKey, []byte{0x00})

	res := getRetByte(0, "invoke endorse success")
	return shim.Success(res)
}

// 老师背书人接受背书
// args: 0 - Bill_No ; 1 - Endorser CmId ; 2 - Endorser Acct
func (a *WorkflowChaincode) accept_teacher(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		res := getRetString(1, "WorkflowChaincode Invoke accept args<3")
		return shim.Error(res)
	}
	// 根据票号取得票据
	bill, bl := a.getBill(stub, args[0])
	if !bl {
		res := getRetString(1, "WorkflowChaincode Invoke accept get bill error")
		return shim.Error(res)
	}

	// 维护search表: 待背书人ID和票号构造复合key 从search表中删除该key
	holderNameBillNoIndexKey, err := stub.CreateCompositeKey(IndexName, []string{bill.WaitEndorserCmID, bill.BillInfoID})
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Invoke accept put search table failed")
		return shim.Error(res)
	}
	stub.DelState(holderNameBillNoIndexKey)

	// 更改票据信息和状态并保存票据: 将前手持票人改为背书人,重置待背书人,票据状态改为背书签收
	bill.WaitEndorserCmID = "0"
	bill.WaitEndorserAcct = "School"
	bill.State = BillInfo_State_TeacherSigned
	// 保存票据
	_, bl = a.putBill(stub, bill)
	if !bl {
		res := getRetString(1, "WorkflowChaincode Invoke accept put bill failed")
		return shim.Error(res)
	}
	// 以待背书人ID和票号构造复合key 向search表中保存 value为空即可 以便待背书人批量查询
	holderNameBillNoIndexKey, err = stub.CreateCompositeKey(IndexName, []string{bill.WaitEndorserCmID, bill.BillInfoID})
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Invoke endorse put search table failed")
		return shim.Error(res)
	}
	stub.PutState(holderNameBillNoIndexKey, []byte{0x00})

	res := getRetByte(0, "invoke accept success")
	return shim.Success(res)
}

// 学校背书人接受背书
// args: 0 - Bill_No ; 1 - Endorser CmId ; 2 - Endorser Acct
func (a *WorkflowChaincode) accept_school(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		res := getRetString(1, "WorkflowChaincode Invoke accept args<3")
		return shim.Error(res)
	}
	// 根据票号取得票据
	bill, bl := a.getBill(stub, args[0])
	if !bl {
		res := getRetString(1, "WorkflowChaincode Invoke accept get bill error")
		return shim.Error(res)
	}

	// 维护search表: 待背书人ID和票号构造复合key 从search表中删除该key
	holderNameBillNoIndexKey, err := stub.CreateCompositeKey(IndexName, []string{bill.WaitEndorserCmID, bill.BillInfoID})
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Invoke accept put search table failed")
		return shim.Error(res)
	}
	stub.DelState(holderNameBillNoIndexKey)
	var return_info string
	//如果逾期
	duetime, err := strconv.Atoi(bill.BillInfoDueDate)
	if bill.BillInfoDueDate != "" && duetime < int(time.Now().Unix()) {
		//状态改为逾期
		bill.WaitEndorserCmID = ""
		bill.WaitEndorserAcct = ""
		bill.State = BillInfo_State_OverDue

		//写入逾期惩罚
		stub.PutState(OverDue_Prefix+bill.DrwrCmID, []byte("overdue"))
		return_info = "This bill is overdue"
	} else {
		// 更改票据信息和状态并保存票据: 将前手持票人改为背书人,重置待背书人,票据状态改为背书签收
		bill.WaitEndorserCmID = ""
		bill.WaitEndorserAcct = ""
		bill.State = BillInfo_State_SchoolSigned
		return_info = "invoke accept success"
	}

	// 保存票据
	_, bl = a.putBill(stub, bill)
	if !bl {
		res := getRetString(1, "WorkflowChaincode Invoke accept put bill failed")
		return shim.Error(res)
	}

	res := getRetByte(0, return_info)
	return shim.Success(res)
}

// 背书人拒绝背书
// args: 0 - Bill_No ; 1 - Endorser CmId, School is "0" ; 2 - Endorser Acct
func (a *WorkflowChaincode) reject(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		res := getRetString(1, "WorkflowChaincode Invoke reject args<3")
		return shim.Error(res)
	}
	// 根据票号取得票据
	bill, bl := a.getBill(stub, args[0])
	if !bl {
		res := getRetString(1, "WorkflowChaincode Invoke reject get bill error")
		return shim.Error(res)
	}

	// 维护search表: 以当前背书人ID和票号构造复合key 从search表中删除该key 以便当前背书人无法再查到该票据
	holderNameBillNoIndexKey, err := stub.CreateCompositeKey(IndexName, []string{args[1], bill.BillInfoID})
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Invoke reject put search table failed")
		return shim.Error(res)
	}
	stub.DelState(holderNameBillNoIndexKey)

	// 更改票据信息和状态并保存票据: 将拒绝背书人改为当前背书人，重置待背书人,票据状态改为背书拒绝
	bill.WaitEndorserCmID = ""
	bill.WaitEndorserAcct = ""
	bill.RejectEndorserCmID = args[1]
	bill.RejectEndorserAcct = args[2]
	if args[1] == "0" {
		bill.State = BillInfo_State_SchoolReject
	} else {
		bill.State = BillInfo_State_TeacherReject
	}

	// 保存票据
	_, bl = a.putBill(stub, bill)
	if !bl {
		res := getRetString(1, "WorkflowChaincode Invoke reject put bill failed")
		return shim.Error(res)
	}

	res := getRetByte(0, "invoke accept success")
	return shim.Success(res)
}

// 查询我的票据:根据持票人编号 批量查询票据
//  0 - Holder CmId ;
func (a *WorkflowChaincode) queryMyBill(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "WorkflowChaincode queryMyBill args!=1")
		return shim.Error(res)
	}
	// 以持票人ID从search表中批量查询所持有的票号
	billsIterator, err := stub.GetStateByPartialCompositeKey(IndexName, []string{args[0]})
	if err != nil {
		res := getRetString(1, "WorkflowChaincode queryMyBill get bill list error")
		return shim.Error(res)
	}
	defer billsIterator.Close()

	var billList = []Bill{}

	for billsIterator.HasNext() {
		kv, _ := billsIterator.Next()
		// 取得持票人名下的票号
		_, compositeKeyParts, err := stub.SplitCompositeKey(kv.Key)
		if err != nil {
			res := getRetString(1, "WorkflowChaincode queryMyBill SplitCompositeKey error")
			return shim.Error(res)
		}
		// 根据票号取得票据
		bill, bl := a.getBill(stub, compositeKeyParts[1])
		if !bl {
			res := getRetString(1, "WorkflowChaincode queryMyBill get bill error")
			return shim.Error(res)
		}
		billList = append(billList, bill)
	}
	// 取得并返回票据数组
	b, err := json.Marshal(billList)
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Marshal queryMyBill billList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// 查询我的待背书票据: 根据背书人编号 批量查询票据
//  0 - Endorser CmId ;
func (a *WorkflowChaincode) queryMyWaitBill(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "WorkflowChaincode queryMyWaitBill args!=1")
		return shim.Error(res)
	}
	// 以背书人ID从search表中批量查询所持有的票号
	billsIterator, err := stub.GetStateByPartialCompositeKey(IndexName, []string{args[0]})
	if err != nil {
		res := getRetString(1, "WorkflowChaincode queryMyWaitBill GetStateByPartialCompositeKey error")
		return shim.Error(res)
	}
	defer billsIterator.Close()

	var billList = []Bill{}

	for billsIterator.HasNext() {
		kv, _ := billsIterator.Next()
		// 从search表中批量查询与背书人有关的票号
		_, compositeKeyParts, err := stub.SplitCompositeKey(kv.Key)
		if err != nil {
			res := getRetString(1, "WorkflowChaincode queryMyWaitBill SplitCompositeKey error")
			return shim.Error(res)
		}
		// 根据票号取得票据
		bill, bl := a.getBill(stub, compositeKeyParts[1])
		if !bl {
			res := getRetString(1, "WorkflowChaincode queryMyWaitBill get bill error")
			return shim.Error(res)
		}
		// 取得状态为待背书的票据 并且待背书人是当前背书人
		if (bill.State == BillInfo_State_WaitTeacherSign || bill.State == BillInfo_State_WaitSchoolSign) && bill.WaitEndorserCmID == args[0] {
			billList = append(billList, bill)
		}
	}

	// 取得并返回票据数组
	b, err := json.Marshal(billList)
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Marshal queryMyWaitBill billList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// 根据票号取得票据 以及该票据背书历史
//  0 - Bill_No ;
func (a *WorkflowChaincode) queryByBillNo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "WorkflowChaincode queryByBillNo args!=1")
		return shim.Error(res)
	}
	// 取得该票据
	bill, bl := a.getBill(stub, args[0])
	if !bl {
		res := getRetString(1, "WorkflowChaincode queryByBillNo get bill error")
		return shim.Error(res)
	}

	// 取得背书历史: 通过fabric api取得该票据的变更历史
	resultsIterator, err := stub.GetHistoryForKey(Bill_Prefix + args[0])
	if err != nil {
		res := getRetString(1, "WorkflowChaincode queryByBillNo GetHistoryForKey error")
		return shim.Error(res)
	}
	defer resultsIterator.Close()

	var history []HistoryItem
	var hisBill Bill
	for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
			res := getRetString(1, "WorkflowChaincode queryByBillNo resultsIterator.Next() error")
			return shim.Error(res)
		}

		var hisItem HistoryItem
		hisItem.TxId = historyData.TxId             //copy transaction id over
		json.Unmarshal(historyData.Value, &hisBill) //un stringify it aka JSON.parse()
		if historyData.Value == nil {               //bill has been deleted
			var emptyBill Bill
			hisItem.Bill = emptyBill //copy nil marble
		} else {
			json.Unmarshal(historyData.Value, &hisBill) //un stringify it aka JSON.parse()
			hisItem.Bill = hisBill                      //copy bill over
		}
		history = append(history, hisItem) //add this tx to the list
	}
	// 将背书历史做为票据的一个属性 一同返回
	bill.History = history

	b, err := json.Marshal(bill)
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Marshal queryByBillNo billList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

// 检查是否超时违约
//  0 - Bill_No ;
func (a *WorkflowChaincode) checkDue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		res := getRetString(1, "WorkflowChaincode queryByBillNo args!=1")
		return shim.Error(res)
	}
	// 取得该票据
	bill, bl := a.getBill(stub, args[0])
	if !bl {
		res := getRetString(1, "WorkflowChaincode queryByBillNo get bill error")
		return shim.Error(res)
	}

	// 取得背书历史: 通过fabric api取得该票据的变更历史
	resultsIterator, err := stub.GetHistoryForKey(Bill_Prefix + args[0])
	if err != nil {
		res := getRetString(1, "WorkflowChaincode queryByBillNo GetHistoryForKey error")
		return shim.Error(res)
	}
	defer resultsIterator.Close()

	var history []HistoryItem
	var hisBill Bill
	for resultsIterator.HasNext() {
		historyData, err := resultsIterator.Next()
		if err != nil {
			res := getRetString(1, "WorkflowChaincode queryByBillNo resultsIterator.Next() error")
			return shim.Error(res)
		}
		var hisItem HistoryItem
		hisItem.TxId = historyData.TxId             //copy transaction id over
		json.Unmarshal(historyData.Value, &hisBill) //un stringify it aka JSON.parse()
		if historyData.Value == nil {               //bill has been deleted
			var emptyBill Bill
			hisItem.Bill = emptyBill //copy nil marble
		} else {
			json.Unmarshal(historyData.Value, &hisBill) //un stringify it aka JSON.parse()
			hisItem.Bill = hisBill                      //copy bill over
		}
		history = append(history, hisItem) //add this tx to the list
	}
	// 将背书历史做为票据的一个属性 一同返回
	bill.History = history

	//检查是否逾期
	duetime, err := strconv.Atoi(bill.BillInfoDueDate)
	if bill.BillInfoDueDate != "" && duetime < int(time.Now().Unix()) {
		//状态改为逾期
		bill.WaitEndorserCmID = ""
		bill.WaitEndorserAcct = ""
		bill.State = BillInfo_State_OverDue
		//写入逾期惩罚
		stub.PutState(OverDue_Prefix+bill.DrwrCmID, []byte("overdue"))
	}
	b, err := json.Marshal(bill)
	if err != nil {
		res := getRetString(1, "WorkflowChaincode Marshal queryByBillNo billList error")
		return shim.Error(res)
	}
	return shim.Success(b)
}

func main() {
	err := shim.Start(new(WorkflowChaincode))
	if err != nil {
		logger.Errorf("Error starting Workflow chaincode: %s", err)
	}
}
