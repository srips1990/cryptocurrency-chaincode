/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"gerrit.hyperledger.org/r/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var marbleIndexStr = "_marbleindex"				//name for the key/value that will store a list of all known marbles
var openTradesStr = "_opentrades"				//name for the key/value that will store all open trades

type UserAccount struct{
	UserId string `json:"userid"`					//the fieldtags are needed to keep case from bouncing around
	AccountNum string `json:"accountnum"`
}

type UserAssets struct{
	UserId string `json:"userid"`					//the fieldtags are needed to keep case from bouncing around
	Assets []Asset `json:"assets"`
}

type Asset struct{
	Name string `json:"name"`					//the fieldtags are needed to keep case from bouncing around
	Qty int `json:"qty"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var str string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	str = args[0]

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(str))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "delete" {										//deletes an entity from its state
		return t.Delete(stub, args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_marble" {									//create a new marble
		return t.init_marble(stub, args)
	} else if function == "set_user" {										//change owner of a marble
		return nil, nil
	} else if function == "transfer_money" {										//change owner of a marble
		return t.transfer_money(stub, args)
	} else if function == "create_user" {										//change owner of a marble
		return t.create_user(stub, args)
	}

	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}


// ============================================================================================================================
// readUserAccountDetails - read User Account Details from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read_user_account_details(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var userId, accountNum, jsonResp string
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	userId = args[0]
	accountNum = args[1]

	accountNumAsbytes, err := stub.GetState(userId)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + userId + "\"}"
		return nil, errors.New(jsonResp)
	}

	if accountNum != string(accountNumAsbytes) {
		return nil, errors.New("This account either doesn't exist or it doesn't belong to the user");
	}

	userAccountDetailsAsbytes, err := stub.GetState(accountNum)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + accountNum + "\"}"
		return nil, errors.New(jsonResp)
	}

	return userAccountDetailsAsbytes, nil													//send it onward
}


// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	name := args[0]
	err := stub.DelState(name)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}


// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Create user - Create user with initial assets
// ============================================================================================================================
func (t *SimpleChaincode) create_user(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var userid, accountNum string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	userid = args[0]															//rename for funsies
	accountNum = args[1]

	valAsbytes, err := stub.GetState(userid)

	if err != nil {
		return nil, err
	}

	if valAsbytes != nil {
		return nil, errors.New("userid already used.")
	}

	err = stub.PutState(userid, []byte(accountNum))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}

	var asset Asset
	asset.Name = "USD";
	asset.Qty = 10000;
	var userAssets UserAssets
	userAssets.UserId = userid
	userAssets.Assets = append(userAssets.Assets, asset)
	userAssetsJsonAsBytes, err := json.Marshal(userAssets)

	if err != nil {
		return nil, err
	}

	err = stub.PutState(accountNum, userAssetsJsonAsBytes)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// transfer_money - transfer money from one account to another
// ============================================================================================================================
func (t *SimpleChaincode) transfer_money(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var fromUserId, toUserId, assetName string // Entities
	var qty int
	var err error
	fmt.Println("running write()")

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	fromUserId = args[0]															//rename for funsies
	toUserId = args[1]
	assetName = args[2]															//rename for funsies
	qty, err = strconv.Atoi(args[3])
	if err != nil {
		return nil, err
	}

	if qty <= 0 {
		return nil, errors.New("Invalid Amount. Amount should be greater than 0")
	}

	senderAccNoAsBytes, err := stub.GetState(fromUserId)		//Getting sender account details
	if err != nil {
		return nil, err
	}

	senderAccNo := string(senderAccNoAsBytes)

	//
	//Sender Assets Validation
	//
	var senderAssets UserAssets
	senderAssetsAsBytes, err := stub.GetState(senderAccNo)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(senderAssetsAsBytes, &senderAssets)
	senderAssetsOld := senderAssets

	var senderAssetIndex int = -1
	for i:=0; i<len(senderAssets.Assets); i++	{
		if senderAssets.Assets[i].Name == assetName {
				senderAssetIndex = i
		}
	}

	if senderAssetIndex == -1 {
		return nil, errors.New("Sender does not hold any unit(s) of this asset")
	}

	//validating asset quantity
	if senderAssets.Assets[senderAssetIndex].Qty < qty {
		return nil, errors.New("Insufficient funds")
	}

	senderAssets.Assets[senderAssetIndex].Qty -= qty

	//
	// Recipient Assets Validation
	//

	recipientAccNoAsBytes, err := stub.GetState(toUserId)		//Getting recipient account details
	if err != nil {
	  return nil, err
	}

	recipientAccNo := string(recipientAccNoAsBytes)		//State Key for Recipient Assets

	var recipientAssets UserAssets		//Goes into PutState
	recipientAssetsAsBytes, err := stub.GetState(recipientAccNo)
	if err != nil {
	  return nil, err
	}
	json.Unmarshal(recipientAssetsAsBytes, &recipientAssets)
	recipientAssetsOld := recipientAssets

	var recipientAssetIndex int = -1
	for i:=0; i<len(recipientAssets.Assets); i++	{
	  if recipientAssets.Assets[i].Name == assetName {
	      recipientAssetIndex = i
	  }
	}

	if recipientAssetIndex == -1 {
	  var recipientAsset Asset
		recipientAsset.Name = assetName
		recipientAsset.Qty = qty
		recipientAssets.Assets = append(recipientAssets.Assets, recipientAsset)
	}	else	{
		recipientAssets.Assets[recipientAssetIndex].Qty -= qty
	}

	//validating asset quantity
	if recipientAssets.Assets[recipientAssetIndex].Qty < qty {
	  return nil, errors.New("Insufficient funds")
	}

	recipientAssetsAsBytes, err := json.Marshal(recipientAssets)
	if err != nil {	return nil, err	}

	senderAssetsAsBytes, err := json.Marshal(senderAssets)
	if err != nil {	return nil, err	}

	recipientAssetsOldAsBytes, err := json.Marshal(recipientAssetsOld)
	if err != nil {	return nil, err	}

	senderAssetsOldAsBytes, err := json.Marshal(senderAssetsOld)
	if err != nil {	return nil, err	}


	err = stub.PutState(recipientAccNo, recipientAssetsAsBytes)					//write the variable into the chaincode state
	if err != nil {	return nil, err	}

	err = stub.PutState(senderAccNo, senderAssetsAsBytes)								//write the variable into the chaincode state
	if err != nil {
		_ = stub.PutState(recipientAccNo, recipientAssetsOldAsBytes)			//rollback on error
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Init Marble - create a new marble, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_marble(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}

	size, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}

	color := strings.ToLower(args[1])
	user := strings.ToLower(args[3])

	str := `{"name": "` + args[0] + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "user": "` + user + `"}`
	err = stub.PutState(args[0], []byte(str))								//store marble with id as key
	if err != nil {
		return nil, err
	}

	//get the marble index
	marblesAsBytes, err := stub.GetState(marbleIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get marble index")
	}
	var marbleIndex []string
	json.Unmarshal(marblesAsBytes, &marbleIndex)							//un stringify it aka JSON.parse()

	//append
	marbleIndex = append(marbleIndex, args[0])								//add marble name to index list
	fmt.Println("! marble index: ", marbleIndex)
	jsonAsBytes, _ := json.Marshal(marbleIndex)
	err = stub.PutState(marbleIndexStr, jsonAsBytes)						//store name of marble

	fmt.Println("- end init marble")
	return nil, nil
}
/*
// ============================================================================================================================
// Set User Permission on Marble
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	marbleAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := UserAssets{}
	json.Unmarshal(marbleAsBytes, &res)										//un stringify it aka JSON.parse()
	res.User = args[1]														//change the user

	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the marble with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end set user")
	return nil, nil
}
*/
