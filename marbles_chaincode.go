package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type work struct {
	ObjectType       string `json:"docType"`          // 状态	
	Uid              string `json:"uid"`              // 用户唯一ID（32位MD5值）
	Workexperience   string `json:"workexperience"`   // 用户工作经历
	WorkStartDate    string `json:"workStartDate"`    // 工作开始日期
	WorkEndDate      string `json:"workEndDate"`      // 工作终止日期
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initwork" { //create a new work
		return t.initwork(stub, args)
	} else if function == "delete" { //delete a work
		return t.delete(stub, args)  
	} else if function == "readwork" { //read a work
		return t.readwork(stub, args)
	} else if function == "queryworks" { //find works based on an ad hoc rich query
		return t.queryworks(stub, args)
	} 
	
	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initwork - create a new work, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initwork(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	// ==== Input sanitation ====
	fmt.Println("- start init work")
	uid           := args[0]
	workexperience:= args[1]
	objectType    := args[2]
	workstartdate := args[3]
	workenddate   := args[4]


	// ==== Create work object and marshal to JSON ====
	objectType := "work"
	work := &work{uid, workexperience, objectType, workstartdate, workenddate}
	workJSONasBytes, err := json.Marshal(work)
	if err != nil {
		return shim.Error(err.Error())
	}


	// === Save work to state ===
	err = stub.PutState(uid, workJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}


	indexName := "workexperience~name"
	workexperienceNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{work.workexperience, work.uid})
	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}
	stub.PutState(workexperienceNameIndexKey, value)

	// ==== work saved and indexed. Return success ====
	fmt.Println("- end init work")
	return shim.Success(nil)
}
/*
// ===============================================
// readwork - read a work from chaincode state
// ===============================================
func (t *SimpleChaincode) readwork(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the work to query")
	}

	name = args[0]
	valJsonBytes, err := stub.GetState(name) //get the work from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valJsonBytes == nil {
		jsonResp = "{\"Error\":\"work does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valJsonBytes)
}

// ==================================================
// delete - remove a work key/value pair from state
// ==================================================
/*
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var workJSON work
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	uid := args[0]

	// to maintain the workexperience~name index, we need to read the work first and get its workexperience
	valJsonBytes, err := stub.GetState(uid) //get the work from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + uid + "\"}"
		return shim.Error(jsonResp)
	} else if valJsonBytes == nil {
		jsonResp = "{\"Error\":\"work does not exist: " + uid + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valJsonBytes), &workJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + uid + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(uid) //remove the work from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "workexperience~name"
	workexperienceNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{workJSON.workexperience, workJSON.uid})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(workexperienceNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// transfer a work by setting a new workstartdate name on the work
// ===========================================================

func (t *SimpleChaincode) queryworks(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}
*/

