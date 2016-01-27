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

	"github.com/openblockchain/obc-peer/openchain/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var marbleIndex []string
var _allIndex []string						//store all variables names here
var _all string = "_all"					//key name for above. tracks all variables in chaincode state

type Marble struct{
	Name string `json:"name"`				//the fieldtags are needed to keep case from bouncing around
	Color string `json:"color"`
	Size int `json:"size"`
	User string `json:"user"`
}

func (t *SimpleChaincode) init(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var Aval int 																	// Asset holdings
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d\n", Aval)

	// Write the state to the ledger
	err = stub.PutState("a", []byte(strconv.Itoa(Aval)))
	if err != nil {
		return nil, err
	}
	
	_allIndex = _allIndex[:0]											//clear the _all index
	jsonAsBytes1, _ := json.Marshal(_allIndex)
	err = stub.PutState(_all, jsonAsBytes1)
	if err != nil {
		return nil, err
	}

	marbleIndex = marbleIndex[:0]										//clear the marble index
	jsonAsBytes, _ := json.Marshal(marbleIndex)
	err = stub.PutState("marbleIndex", jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Run
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)

	// Handle different functions
	if function == "init" {													// Initialize the entities and their asset holdings
		return t.init(stub, args)
	} else if function == "delete" {										// Deletes an entity from its state
		return t.Delete(stub, args)
	} else if function == "write" {											// Writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_marble" {									//init_marble
		return t.init_marble(stub, args)
	} else if function == "set_user" {										//set user permissions
		return t.set_user(stub, args)
	}
	fmt.Println("run issues " + function)

	return nil, errors.New("Received unknown function invocation")
}

// Deletes an entity from state
func (t *SimpleChaincode) Delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	name := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(name)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	for i,val := range marbleIndex{
		fmt.Println(strconv.Itoa(i) + "looking at " + val + " for " + name)
		if val == name{															//find the correct marble
			fmt.Println("found it")
			marbleIndex = append(marbleIndex[:i], marbleIndex[i+1:]...)			//remove it
			break
		}
	}
	
	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var name string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	name = args[0]

	// Get the state from the ledger
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"" + name + "\":\"" + string(valAsbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return []byte(jsonResp), nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//----------------------------------------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------------------------------------

// ============================================================================================================================
// Write var into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write - start")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the variable and value to set")
	}

	name = args[0]
	value = args[1]

	// Write the state back to the ledger
	err = stub.PutState(name, []byte(value))
	if err != nil {
		return nil, err
	}
	t.remember_me(stub, name)

	fmt.Println("running write - fin - for name " + name)
	return nil, nil
}

// ============================================================================================================================
// Remember Me - remember the name of variables we stored in ledger 
// ============================================================================================================================
func (t *SimpleChaincode) remember_me(stub *shim.ChaincodeStub, name string) ([]byte, error) {		//dsh - to do, should probably not exist here, move to stub
	var err error
	/*var storedNames string
	storeNamesAsBytes, err := stub.GetState(_all)
	if err != nil {
		return nil, errors.New("Failed to get _all")
	}

	storedNames = string(storeNamesAsBytes)
	// Write the state back to the ledger
	err = stub.PutState(_all, []byte(storedNames + "," + name))										//dsh - to do, should probably be json
	if err != nil {
		return nil, err
	}
	*/
	
	_allIndex = append(_allIndex, name)									//add var name to index list
	jsonAsBytes, _ := json.Marshal(_allIndex)
	err = stub.PutState(_all, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Init Person 
// ============================================================================================================================
func (t *SimpleChaincode) init_marble(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	val, err := strconv.Atoi(args[2])
	if err != nil {
		val = 16														//default value
	}
	str := `{"name": "` + args[0] + `", "color": "` + args[1] + `", "size": "` + strconv.Itoa(val) + `", "user": "` + args[3] + `"}`

	// Write the state back to the ledger
	err = stub.PutState(args[0], []byte(str))							//store marble with id as key
	if err != nil {
		return nil, err
	}
	t.remember_me(stub, args[0])
	
	marbleIndex = append(marbleIndex, args[0])							//add marble name to index list
	jsonAsBytes, _ := json.Marshal(marbleIndex)
	err = stub.PutState("marbleIndex", jsonAsBytes)						//store name of marble

	return nil, nil
}

// ============================================================================================================================
// Set User Permissions 
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	
	marbleAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Marble{}
	json.Unmarshal(marbleAsBytes, &res)
	fmt.Println(res)
	
	/*for i,perm := range res.Users{
		if perm.UserId == args[1]{							//find the correct user
			res.Users[i].Permissions[0] = args[2]			//set new perm, dsh - to do make this input as array of all perms
			fmt.Println(res.Users[i].Permissions)
			break
		}
	}*/
	res.User = args[1]
	
	// Write the state back to the ledger
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}