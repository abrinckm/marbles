/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright messengership.  The ASF licenses this file
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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ============================================================================================================================
// Read - read a generic variable from ledger
//
// Shows Off GetState() - reading a key/value from the ledger
//
// Inputs - Array of strings
//  0
//  key
//  "abc"
//
// Returns - string
// ============================================================================================================================
func read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, jsonResp string
	var err error
	fmt.Println("starting read")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting key of the var to query")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)           //get the var from ledger
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("- end read")
	return shim.Success(valAsbytes)                  //send it onward
}

// ============================================================================================================================
// Get everything we need (messengers + messages + companies)
//
// Inputs - none
//
// Returns:
// {
//	"messengers": [{
//			"id": "o99999999",
//			"username": "alice"
//	}],
//	"messages": [{
//		"id": "m1490898165086",
//		"text": "hello",
//		"docType" :"message",
//		"messenger": {
//			"username": "alice"
//		},
//		"recipient" : "ava"
//	}]
// }
// ============================================================================================================================
func read_everything(stub shim.ChaincodeStubInterface) pb.Response {
	type Everything struct {
		Messengers   []Messenger   `json:"messengers"`
		Messages  []Message  `json:"messages"`
	}
	var everything Everything

	// ---- Get All Messages ---- //
	resultsIterator, err := stub.GetStateByRange("m0", "m9999999999999999999")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryKeyAsStr, queryValAsBytes, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		fmt.Println("on message id - ", queryKeyAsStr)
		var message Message
		json.Unmarshal(queryValAsBytes, &message)                  //un stringify it aka JSON.parse()
		everything.Messages = append(everything.Messages, message)   //add this message to the list
	}
	fmt.Println("message array - ", everything.Messages)

	// ---- Get All Messengers ---- //
	messengersIterator, err := stub.GetStateByRange("o0", "o9999999999999999999")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer messengersIterator.Close()

	for messengersIterator.HasNext() {
		queryKeyAsStr, queryValAsBytes, err := messengersIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		fmt.Println("on messenger id - ", queryKeyAsStr)
		var messenger Messenger
		json.Unmarshal(queryValAsBytes, &messenger)                  //un stringify it aka JSON.parse()
		everything.Messengers = append(everything.Messengers, messenger)     //add this message to the list
	}
	fmt.Println("messenger array - ", everything.Messengers)

	//change to array of bytes
	everythingAsBytes, _ := json.Marshal(everything)             //convert to array of bytes
	return shim.Success(everythingAsBytes)
}

// ============================================================================================================================
// Get history of asset
//
// Shows Off GetHistoryForKey() - reading complete history of a key/value
//
// Inputs - Array of strings
//  0
//  id
//  "m01490985296352SjAyM"
// ============================================================================================================================
func getHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	type AuditHistory struct {
		TxId    string   `json:"txId"`
		Value   Message   `json:"value"`
	}
	var history []AuditHistory;
	var message Message

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	messageId := args[0]
	fmt.Printf("- start getHistoryForMessage: %s\n", messageId)

	// Get History
	resultsIterator, err := stub.GetHistoryForKey(messageId)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		txID, historicValue, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var tx AuditHistory
		tx.TxId = txID                             //copy transaction id over
		json.Unmarshal(historicValue, &message)     //un stringify it aka JSON.parse()
		if historicValue == nil {                  //message has been deleted
			var emptyMessage Message
			tx.Value = emptyMessage                 //copy nil message
		} else {
			json.Unmarshal(historicValue, &message) //un stringify it aka JSON.parse()
			tx.Value = message                      //copy message over
		}
		history = append(history, tx)              //add this tx to the list
	}
	fmt.Printf("- getHistoryForMessage returning:\n%s", history)

	//change to array of bytes
	historyAsBytes, _ := json.Marshal(history)     //convert to array of bytes
	return shim.Success(historyAsBytes)
}

// ============================================================================================================================
// Get history of asset - performs a range query based on the start and end keys provided.
//
// Shows Off GetStateByRange() - reading a multiple key/values from the ledger
//
// Inputs - Array of strings
//       0     ,    1
//   startKey  ,  endKey
//  "messages1" , "messages5"
// ============================================================================================================================
func getMessagesByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
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
		queryResultKey, queryResultValue, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResultKey)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResultValue))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getMessagesByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
