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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ============================================================================================================================
// write() - genric write variable into ledger
//
// Shows Off PutState() - writting a key/value into the ledger
//
// Inputs - Array of strings
//    0   ,    1
//   key  ,  value
//  "abc" , "test"
// ============================================================================================================================
func write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string
	var err error
	fmt.Println("starting write")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0]                                   //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value))         //write the variable into the ledger
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end write")
	return shim.Success(nil)
}

// ============================================================================================================================
// delete_message() - remove a message from state and from message index
//
// Shows Off DelState() - "removing"" a key/value from the ledger
//
// Inputs - Array of strings
//      0      ,         1
//     id      ,  authed_by_company
// "m999999999", "united messages"
// ============================================================================================================================
func delete_message(stub shim.ChaincodeStubInterface, args []string) (pb.Response) {
	fmt.Println("starting delete_message")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// input sanitation
	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	authed_by_company := args[1]

	// get the message
	message, err := get_message(stub, id)
	if err != nil{
		fmt.Println("Failed to find message by id " + id)
		return shim.Error(err.Error())
	}

	// check authorizing company (see note in set_messenger() about how this is quirky)
	if message.Messenger.Company != authed_by_company{
		return shim.Error("The company '" + authed_by_company + "' cannot authorize deletion for '" + message.Messenger.Company + "'.")
	}

	// remove the message
	err = stub.DelState(id)                                                 //remove the key from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("- end delete_message")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init Message - create a new message, store into chaincode state
//
// Shows off building a key's JSON value manually
//
// Inputs - Array of strings
//      0      ,    1  ,      2  ,           3         ,       4
//     id      ,  text, ,  priority ,  messenger id    ,  recipient id
// "m999999999", "hi !",   "1",      "o9999999999999", "o9999999999999"
// ============================================================================================================================
func init_message(stub shim.ChaincodeStubInterface, args []string) (pb.Response) {
	var err error
	fmt.Println("starting init_message")

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	id := args[0]
	text := args[1]
	messenger_id := args[3]
	// recipient_id := args[4]
	priority, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	//check if new messenger exists
	messenger, err := get_messenger(stub, messenger_id)
	if err != nil {
		fmt.Println("Failed to find messenger - " + messenger_id)
		return shim.Error(err.Error())
	}

	//check authorizing company (see note in set_messenger() about how this is quirky)
	// if messenger.Company != authed_by_company{
	// 	return shim.Error("The company '" + authed_by_company + "' cannot authorize creation for '" + messenger.Company + "'.")
	// }

	//check if message id already exists
	message, err := get_message(stub, id)
	if err == nil {
		fmt.Println("This message already exists - " + id)
		fmt.Println(message)
		return shim.Error("This message already exists - " + id)  //all stop a message by this id exists
	}

	//build the message json string manually
	// "company": "` + messenger.Company + `"
	str := `{
		"docType":"message",
		"id": "` + id + `",
		"text": "` + text + `",
		"priority": ` + strconv.Itoa(priority) + `,
		"messenger": {
			"id": "` + messenger_id + `",
			"username": "` + messenger.Username + `"
		}
	}`
	err = stub.PutState(id, []byte(str))                         //store message with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_message")
	return shim.Success(nil)
}

// ============================================================================================================================
// Init Messenger - create a new messenger aka end user, store into chaincode state
//
// Shows off building key's value from GoLang Structure
//
// Inputs - Array of Strings
//           0     ,     1   ,   2
//      messenger id   , username, company
// "o9999999999999",     bob", "united messages"
// ============================================================================================================================
func init_messenger(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting init_messenger")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	var messenger Messenger
	messenger.ObjectType = "message_messenger"
	messenger.Id =  args[0]
	messenger.Username = strings.ToLower(args[1])
	// messenger.Company = args[2]
	fmt.Println(messenger)

	//check if user already exists
	_, err = get_messenger(stub, messenger.Id)
	if err == nil {
		fmt.Println("This messenger already exists - " + messenger.Id)
		return shim.Error("This messenger already exists - " + messenger.Id)
	}

	//store user
	messengerAsBytes, _ := json.Marshal(messenger)                         //convert to array of bytes
	err = stub.PutState(messenger.Id, messengerAsBytes)                    //store messenger by its Id
	if err != nil {
		fmt.Println("Could not store user")
		return shim.Error(err.Error())
	}

	fmt.Println("- end init_messenger message")
	return shim.Success(nil)
}

// ============================================================================================================================
// Set Messenger on Message
//
// Shows off GetState() and PutState()
//
// Inputs - Array of Strings
//       0     ,        1      ,        2
//  message id  ,  to messenger id  , company that auth the transfer
// "m999999999", "o99999999999", united_mables"
// ============================================================================================================================
// func set_messenger(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	var err error
// 	fmt.Println("starting set_messenger")
//
// 	// this is quirky
// 	// todo - get the "company that authed the transfer" from the certificate instead of an argument
// 	// should be possible since we can now add attributes to the enrollment cert
// 	// as is.. this is a bit broken (security wise), but it's much much easier to demo! holding off for demos sake
//
// 	if len(args) != 3 {
// 		return shim.Error("Incorrect number of arguments. Expecting 3")
// 	}
//
// 	// input sanitation
// 	err = sanitize_arguments(args)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
//
// 	var message_id = args[0]
// 	var new_messenger_id = args[1]
// 	var authed_by_company = args[2]
// 	fmt.Println(message_id + "->" + new_messenger_id + " - |" + authed_by_company)
//
// 	// check if user already exists
// 	messenger, err := get_messenger(stub, new_messenger_id)
// 	if err != nil {
// 		return shim.Error("This messenger does not exist - " + new_messenger_id)
// 	}
//
// 	// get message's current state
// 	messageAsBytes, err := stub.GetState(message_id)
// 	if err != nil {
// 		return shim.Error("Failed to get message")
// 	}
// 	res := Message{}
// 	json.Unmarshal(messageAsBytes, &res)           //un stringify it aka JSON.parse()
//
// 	// check authorizing company
// 	if res.Messenger.Company != authed_by_company{
// 		return shim.Error("The company '" + authed_by_company + "' cannot authorize transfers for '" + res.Messenger.Company + "'.")
// 	}
//
// 	// transfer the message
// 	res.Messenger.Id = new_messenger_id                   //change the messenger
// 	res.Messenger.Username = messenger.Username
// 	res.Messenger.Company = messenger.Company
// 	jsonAsBytes, _ := json.Marshal(res)           //convert to array of bytes
// 	err = stub.PutState(args[0], jsonAsBytes)     //rewrite the message with id as key
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
//
// 	fmt.Println("- end set messenger")
// 	return shim.Success(nil)
// }
