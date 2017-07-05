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
	"errors"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ============================================================================================================================
// Get Message - get a message asset from ledger
// ============================================================================================================================
func get_message(stub shim.ChaincodeStubInterface, id string) (Message, error) {
	var message Message
	messageAsBytes, err := stub.GetState(id)                  //getState retreives a key/value from the ledger
	if err != nil {                                          //this seems to always succeed, even if key didn't exist
		return message, errors.New("Failed to find message - " + id)
	}
	json.Unmarshal(messageAsBytes, &message)                   //un stringify it aka JSON.parse()

	if message.Id != id {                                     //test if message is actually here or just nil
		return message, errors.New("Message does not exist - " + id)
	}

	return message, nil
}

// ============================================================================================================================
// Get Messenger - get the messenger asset from ledger
// ============================================================================================================================
func get_messenger(stub shim.ChaincodeStubInterface, id string) (Messenger, error) {
	var messenger Messenger
	messengerAsBytes, err := stub.GetState(id)                     //getState retreives a key/value from the ledger
	if err != nil {                                            //this seems to always succeed, even if key didn't exist
		return messenger, errors.New("Failed to get messenger - " + id)
	}
	json.Unmarshal(messengerAsBytes, &messenger)                       //un stringify it aka JSON.parse()

	if len(messenger.Username) == 0 {                              //test if messenger is actually here or just nil
		return messenger, errors.New("Messenger does not exist - " + id + ", '" + messenger.Username + "' '" + messenger.Company + "'")
	}

	return messenger, nil
}

// ========================================================
// Input Sanitation - dumb input checking, look for empty strings
// ========================================================
func sanitize_arguments(strs []string) error{
	for i, val:= range strs {
		if len(val) <= 0 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be a non-empty string")
		}
		if len(val) > 32 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be <= 32 characters")
		}
	}
	return nil
}
