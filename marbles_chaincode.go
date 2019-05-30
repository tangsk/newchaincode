
// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke works ====
// peer chaincode invoke -C myc1 -n works -c '{"Args":["initWork","work1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n works -c '{"Args":["initWork","work2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n works -c '{"Args":["initWork","work3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n works -c '{"Args":["transferWork","work2","jerry"]}'
// peer chaincode invoke -C myc1 -n works -c '{"Args":["transferWorksBasedOnWorkstartdate","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n works -c '{"Args":["delete","work1"]}'

// ==== Query works ====
// peer chaincode query -C myc1 -n works -c '{"Args":["readWork","work1"]}'
// peer chaincode query -C myc1 -n works -c '{"Args":["getWorksByRange","work1","work3"]}'
// peer chaincode query -C myc1 -n works -c '{"Args":["getHistoryForWork","work1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n works -c '{"Args":["queryWorksByWorkexperience","tom"]}'
//   peer chaincode query -C myc1 -n works -c '{"Args":["queryWorks","{\"selector\":{\"workexperience\":\"tom\"}}"]}'

//The following examples demonstrate creating indexes on CouchDB
//Example hostuid:port configurations
//
//Docker or vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for chaincodeid, docType, workexperience.
// Note that docType and workexperience fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":["chaincodeid","data.docType","data.workexperience"]},"ddoc":"indexWorkexperienceDoc", "uid":"indexWorkexperience","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"chaincodeid\",\"data.docType\",\"data.workexperience\"]},\"uid\":\"indexWorkexperience\",\"ddoc\":\"indexWorkexperienceDoc\",\"type\":\"json\"}" http://hostuid:port/myc1/_index
//

// Index for chaincodeid, docType, workexperience, workenddate (descending order).
// Note that docType, workexperience and workenddate fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":[{"data.workenddate":"desc"},{"chaincodeid":"desc"},{"data.docType":"desc"},{"data.workexperience":"desc"}]},"ddoc":"indexWorkenddateSortDoc", "uid":"indexWorkenddateSortDesc","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"data.workenddate\":\"desc\"},{\"chaincodeid\":\"desc\"},{\"data.docType\":\"desc\"},{\"data.workexperience\":\"desc\"}]},\"ddoc\":\"indexWorkenddateSortDoc\", \"uid\":\"indexWorkenddateSortDesc\",\"type\":\"json\"}" http://hostuid:port/myc1/_index

// Rich Query with index design doc and index uid specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n works -c '{"Args":["queryWorks","{\"selector\":{\"docType\":\"work\",\"workexperience\":\"tom\"}, \"use_index\":[\"_design/indexWorkexperienceDoc\", \"indexWorkexperience\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n works -c '{"Args":["queryWorks","{\"selector\":{\"docType\":{\"$eq\":\"work\"},\"workexperience\":{\"$eq\":\"tom\"},\"workenddate\":{\"$gt\":0}},\"fields\":[\"docType\",\"workexperience\",\"workenddate\"],\"sort\":[{\"workenddate\":\"desc\"}],\"use_index\":\"_design/indexWorkenddateSortDoc\"}"]}'

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
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Uid       string `json:"uid"`    //the fieldtags are needed to keep case from bouncing around
	Workstartdate      string `json:"workstartdate"`
	Workenddate       int    `json:"workenddate"`
	Workexperience      string `json:"workexperience"`
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

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initWork" { //create a new work
		return t.initWork(stub, args)
	} else if function == "transferWork" { //change workexperience of a specific work
		return t.transferWork(stub, args)
	} else if function == "transferWorksBasedOnWorkstartdate" { //transfer all works of a certain workstartdate
		return t.transferWorksBasedOnWorkstartdate(stub, args)
	} else if function == "delete" { //delete a work
		return t.delete(stub, args)
	} else if function == "readWork" { //read a work
		return t.readWork(stub, args)
	} else if function == "queryWorksByWorkexperience" { //find works for workexperience X using rich query
		return t.queryWorksByWorkexperience(stub, args)
	} else if function == "queryWorks" { //find works based on an ad hoc rich query
		return t.queryWorks(stub, args)
	} else if function == "getHistoryForWork" { //get history of values for a work
		return t.getHistoryForWork(stub, args)
	} else if function == "getWorksByRange" { //get works based on range query
		return t.getWorksByRange(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initWork - create a new work, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initWork(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init work")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	workUid := args[0]
	workstartdate := strings.ToLower(args[1])
	workexperience := strings.ToLower(args[3])
	workenddate, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if work already exists ====
	workAsBytes, err := stub.GetState(workUid)
	if err != nil {
		return shim.Error("Failed to get work: " + err.Error())
	} else if workAsBytes != nil {
		fmt.Println("This work already exists: " + workUid)
		return shim.Error("This work already exists: " + workUid)
	}

	// ==== Create work object and marshal to JSON ====
	objectType := "work"
	work := &work{objectType, workUid, workstartdate, workenddate, workexperience}
	workJSONasBytes, err := json.Marshal(work)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the work json string manually if you don't want to use struct marshalling
	//workJSONasString := `{"docType":"Work",  "uid": "` + workUid + `", "workstartdate": "` + workstartdate + `", "workenddate": ` + strconv.Itoa(workenddate) + `, "workexperience": "` + workexperience + `"}`
	//workJSONasBytes := []byte(str)

	// === Save work to state ===
	err = stub.PutState(workUid, workJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the work to enable workstartdate-based range queries, e.g. return all blue works ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexUid~workstartdate~uid.
	//  This will enable very efficient state range queries based on composite keys matching indexUid~workstartdate~*
	indexUid := "workstartdate~uid"
	workstartdateUidIndexKey, err := stub.CreateCompositeKey(indexUid, []string{work.Workstartdate, work.Uid})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key uid is needed, no need to store a duplicate copy of the work.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(workstartdateUidIndexKey, value)

	// ==== Work saved and indexed. Return success ====
	fmt.Println("- end init work")
	return shim.Success(nil)
}

// ===============================================
// readWork - read a work from chaincode state
// ===============================================
func (t *SimpleChaincode) readWork(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var uid, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting uid of the work to query")
	}

	uid = args[0]
	valAsbytes, err := stub.GetState(uid) //get the work from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + uid + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Work does not exist: " + uid + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
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
	workUid := args[0]

	// to maintain the workstartdate~uid index, we need to read the work first and get its workstartdate
	valAsbytes, err := stub.GetState(workUid) //get the work from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + workUid + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Work does not exist: " + workUid + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &workJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + workUid + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(workUid) //remove the work from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexUid := "workstartdate~uid"
	workstartdateUidIndexKey, err := stub.CreateCompositeKey(indexUid, []string{workJSON.Workstartdate, workJSON.Uid})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(workstartdateUidIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// transfer a work by setting a new workexperience uid on the work
// ===========================================================
func (t *SimpleChaincode) transferWork(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "uid", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	workUid := args[0]
	newWorkexperience := strings.ToLower(args[1])
	fmt.Println("- start transferWork ", workUid, newWorkexperience)

	workAsBytes, err := stub.GetState(workUid)
	if err != nil {
		return shim.Error("Failed to get work:" + err.Error())
	} else if workAsBytes == nil {
		return shim.Error("Work does not exist")
	}

	workToTransfer := work{}
	err = json.Unmarshal(workAsBytes, &workToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	workToTransfer.Workexperience = newWorkexperience //change the workexperience

	workJSONasBytes, _ := json.Marshal(workToTransfer)
	err = stub.PutState(workUid, workJSONasBytes) //rewrite the work
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferWork (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// getWorksByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getWorksByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

	fmt.Printf("- getWorksByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// transferWorksBasedOnWorkstartdate will transfer works of a given workstartdate to a certain new workexperience.
// Uses a GetStateByPartialCompositeKey (range query) against workstartdate~uid 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) transferWorksBasedOnWorkstartdate(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "workstartdate", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	workstartdate := args[0]
	newWorkexperience := strings.ToLower(args[1])
	fmt.Println("- start transferWorksBasedOnWorkstartdate ", workstartdate, newWorkexperience)

	// Query the workstartdate~uid index by workstartdate
	// This will execute a key range query on all keys starting with 'workstartdate'
	workstartdateedWorkResultsIterator, err := stub.GetStateByPartialCompositeKey("workstartdate~uid", []string{workstartdate})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer workstartdateedWorkResultsIterator.Close()

	// Iterate through result set and for each work found, transfer to newWorkexperience
	var i int
	for i = 0; workstartdateedWorkResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the work uid from the composite key
		responseRange, err := workstartdateedWorkResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the workstartdate and uid from workstartdate~uid composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedWorkstartdate := compositeKeyParts[0]
		returnedWorkUid := compositeKeyParts[1]
		fmt.Printf("- found a work from index:%s workstartdate:%s uid:%s\n", objectType, returnedWorkstartdate, returnedWorkUid)

		// Now call the transfer function for the found work.
		// Re-use the same function that is used to transfer individual works
		response := t.transferWork(stub, []string{returnedWorkUid, newWorkexperience})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
	}

	responsePayload := fmt.Sprintf("Transferred %d %s works to %s", i, workstartdate, newWorkexperience)
	fmt.Println("- end transferWorksBasedOnWorkstartdate: " + responsePayload)
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
// queryWorksByWorkexperience queries for works based on a passed in workexperience.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (workexperience).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryWorksByWorkexperience(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	workexperience := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"work\",\"workexperience\":\"%s\"}}", workexperience)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryWorks uses a query string to perform a query for works.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryWorksForWorkexperience example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryWorks(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

func (t *SimpleChaincode) getHistoryForWork(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	workUid := args[0]

	fmt.Printf("- start getHistoryForWork: %s\n", workUid)

	resultsIterator, err := stub.GetHistoryForKey(workUid)
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

	fmt.Printf("- getHistoryForWork returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
