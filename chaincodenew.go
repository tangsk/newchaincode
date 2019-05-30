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
	ObjectType       string `json:"docType"`          //docType is used to distinguish the various types of objects in state database	
	Uid              string `json:"uid"`              // 用户唯一ID（32位MD5值）
	Workexperience   string `json:"workexperience"`   // 用户工作经历
	ApplyDate        string `json:"applyDate"`        // 申请日期
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
//函数有问题，需要改
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initwork" { //create a new work
		return t.initwork(stub, args)
	} else if function == "transferwork" { //change owner of a specific work
		return t.transferwork(stub, args)
	} else if function == "transferworksBasedOnColor" { //transfer all works of a certain color
		return t.transferworksBasedOnColor(stub, args)
	} else if function == "delete" { //delete a work
		return t.delete(stub, args)
	} else if function == "readwork" { //read a work
		return t.readwork(stub, args)
	} else if function == "queryworksByOwner" { //find works for  X using rich query
		return t.queryworksByOwner(stub, args)
	} else if function == "queryworks" { //find works based on an ad hoc rich query
		return t.queryworks(stub, args)
	} else if function == "getHistoryForwork" { //get history of values for a work
		return t.getHistoryForwork(stub, args)
	} else if function == "getworksByRange" { //get works based on range query
		return t.getworksByRange(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initwork - create a new work, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initwork(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error


	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init work")
	if len(args[0]) != 32 {
		return fmt.Errorf("Parameter uid length error while Work, 32 is right")
	}
	if len(args[2]) != 14 {
		return fmt.Errorf("Parameter ApplyDate length error while Work, 14 is right")
	}
	if len(args[3]) != 14 {
		return fmt.Errorf("Parameter WorkStartDate length error while Work, 14 is right")
	}
	if len(args[4]) != 14 {
		return fmt.Errorf("Parameter WorkEndDate length error while Work, 14 is right")
	}
	uid           := args[0]
	workexperience:= args[1]
	applydate     := args[2]
	workstartdate := args[3]
	workenddate   := args[4]
	key, err := strconv.Atoi(args[5])
	if err != nil {
		return shim.Error("Json serialize Work fail while Work, work id = " + args[5])
	}

	

	// ==== Check if work already exists ====
	workJsonBytes, err := stub.GetState(uid)
	if err != nil {
		return shim.Error("Failed to get work: " + err.Error())
	} else if workJsonBytes != nil {
		fmt.Println("This work already exists: " + uid)
		return shim.Error("This work already exists: " + uid)
	}

	// ==== Create work object and marshal to JSON ====
	objectType := "work"
	work := &work{objectType, uid, color, key, owner}
	workJSONJsonBytes, err := json.Marshal(work)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the work json string manually if you don't want to use struct marshalling
	//workJSONasString := `{"docType":"work",  "name": "` + uid + `", "color": "` + color + `", "key": ` + strconv.Itoa(key) + `, "owner": "` + owner + `"}`
	//workJSONJsonBytes := []byte(str)

	// === Save work to state ===
	err = stub.PutState(uid, workJSONJsonBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the work to enable color-based range queries, e.g. return all blue works ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{work.Color, work.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the work.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)

	// ==== work saved and indexed. Return success ====
	fmt.Println("- end init work")
	return shim.Success(nil)
}

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
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var workJSON work
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	uid := args[0]

	// to maintain the color~name index, we need to read the work first and get its color
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
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{workJSON.Color, workJSON.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(colorNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// transfer a work by setting a new owner name on the work
// ===========================================================
func (t *SimpleChaincode) transferwork(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	uid := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferwork ", uid, newOwner)

	workJsonBytes, err := stub.GetState(uid)
	if err != nil {
		return shim.Error("Failed to get work:" + err.Error())
	} else if workJsonBytes == nil {
		return shim.Error("work does not exist")
	}

	workToTransfer := work{}
	err = json.Unmarshal(workJsonBytes, &workToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	workToTransfer.Owner = newOwner //change the owner

	workJSONJsonBytes, _ := json.Marshal(workToTransfer)
	err = stub.PutState(uid, workJSONJsonBytes) //rewrite the work
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferwork (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// getworksByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getworksByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getworksByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// transferworksBasedOnColor will transfer works of a given color to a certain new owner.
// Uses a GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) transferworksBasedOnColor(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "color", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	color := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferworksBasedOnColor ", color, newOwner)

	// Query the color~name index by color
	// This will execute a key range query on all keys starting with 'color'
	coloredworkResultsIterator, err := stub.GetStateByPartialCompositeKey("color~name", []string{color})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer coloredworkResultsIterator.Close()

	// Iterate through result set and for each work found, transfer to newOwner
	var i int
	for i = 0; coloredworkResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the work name from the composite key
		responseRange, err := coloredworkResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the color and name from color~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedColor := compositeKeyParts[0]
		returneduid := compositeKeyParts[1]
		fmt.Printf("- found a work from index:%s color:%s name:%s\n", objectType, returnedColor, returneduid)

		// Now call the transfer function for the found work.
		// Re-use the same function that is used to transfer individual works
		response := t.transferwork(stub, []string{returneduid, newOwner})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
	}

	responsePayload := fmt.Sprintf("Transferred %d %s works to %s", i, color, newOwner)
	fmt.Println("- end transferworksBasedOnColor: " + responsePayload)
	return shim.Success([]byte(responsePayload))
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryworksByOwner queries for works based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryworksByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"work\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryworks uses a query string to perform a query for works.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryworksForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
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

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) getHistoryForwork(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	uid := args[0]

	fmt.Printf("- start getHistoryForwork: %s\n", uid)

	resultsIterator, err := stub.GetHistoryForKey(uid)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the work
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON work)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForwork returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
