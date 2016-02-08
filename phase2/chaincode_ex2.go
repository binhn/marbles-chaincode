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
	"time"

	"github.com/openblockchain/obc-peer/openchain/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var marbleIndex []string						//store marbles here, b/c we need this
var _allIndex []string							//store all variables names here, b/c its handy for debug
var _all string = "_all"						//key name for above. tracks all variables in chaincode state

type Marble struct{
	Name string `json:"name"`					//the fieldtags are needed to keep case from bouncing around
	Color string `json:"color"`
	Size int `json:"size"`
	User string `json:"user"`
}

type Description struct{
	Color string `json:"color"`
	Size int `json:"size"`
}

type AnOpenTrade struct{
	User string `json:"user"`					//user who created the open trade order
	Timestamp int64 `json:"timestamp"`			//utc timestamp of creation
	Want Description  `json:"want"`				//description of desired marble
	Willing []Description `json:"willing"`		//array of marbles willing to trade away
}

type AllTrades struct{
	OpenTrades []AnOpenTrade `json:"open_trades"`
}

var trades AllTrades

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) init(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var Aval int
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
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	_allIndex = _allIndex[:0]											//clear the _all variables index
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
// Run - Our entry point
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.init(stub, args)
	} else if function == "delete" {										//deletes an entity from its state
		return t.Delete(stub, args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_marble" {									//create a new marble
		return t.init_marble(stub, args)
	} else if function == "set_user" {										//change owner of a marble
		return t.set_user(stub, args)
	} else if function == "open_trade" {									//create a new trade order
		return t.open_trade(stub, args)
	} else if function == "perform_trade" {									//forfill an open trade order
		return t.perform_trade(stub, args)
	}
	fmt.Println("run did not find func: " + function)						//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Delete - remove an entity from state
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

	//remove marble from index
	for i,val := range marbleIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + name)
		if val == name{															//find the correct marble
			fmt.Println("found marble")
			marbleIndex = append(marbleIndex[:i], marbleIndex[i+1:]...)			//remove it
			for x:= range marbleIndex{											//debug prints...
				fmt.Println(string(x) + " - " + marbleIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(marbleIndex)									//save new index
	err = stub.PutState("marbleIndex", jsonAsBytes)

	return nil, nil
}

// ============================================================================================================================
// Query - read a variable from chaincode state - (aka read)
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
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
	t.remember_me(stub, name)

	return nil, nil
}

// ============================================================================================================================
// Remember Me - remember the name of variables we stored in ledger 
// ============================================================================================================================
func (t *SimpleChaincode) remember_me(stub *shim.ChaincodeStub, name string) ([]byte, error) {
	var err error

	_allIndex = append(_allIndex, name)										//add var name to index list
	jsonAsBytes, _ := json.Marshal(_allIndex)
	err = stub.PutState(_all, jsonAsBytes)									//store it
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Init Marble 
// ============================================================================================================================
func (t *SimpleChaincode) init_marble(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	val, err := strconv.Atoi(args[2])
	if err != nil {
		val = 16															//default value
	}
	str := `{"name": "` + args[0] + `", "color": "` + args[1] + `", "size": ` + strconv.Itoa(val) + `, "user": "` + args[3] + `"}`

	err = stub.PutState(args[0], []byte(str))								//store marble with id as key
	if err != nil {
		return nil, err
	}
	t.remember_me(stub, args[0])
	
	marbleIndex = append(marbleIndex, args[0])								//add marble name to index list
	jsonAsBytes, _ := json.Marshal(marbleIndex)
	err = stub.PutState("marbleIndex", jsonAsBytes)							//store name of marble

	return nil, nil
}

// ============================================================================================================================
// Set User Permission on Marble
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	
	// "asdf", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	marbleAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Marble{}
	json.Unmarshal(marbleAsBytes, &res)										//un stringify it aka JSON.parse()
	fmt.Println(res)
	res.User = args[1]														//change the user
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the marble with id as key
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Open Trade - create an open trade for a marble you want with marbles you have 
// ============================================================================================================================
func (t *SimpleChaincode) open_trade(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//"bob", "blue", "16", "red", "16"
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 5?")
	}

	size1, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}
	size2, err := strconv.Atoi(args[4])
	if err != nil {
		return nil, errors.New("5th argument must be a numeric string")
	}

	open := AnOpenTrade{};
	open.User = args[0];
	open.Timestamp = time.Now().UnixNano();										//use this as an ID
	open.Want.Color = args[1];
	open.Want.Size =  size1;
	fmt.Println("! start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	
	trade_away := Description{};
	trade_away.Color = args[3];
	trade_away.Size =  size2;
	fmt.Println("! created trade_away")
	jsonAsBytes, _ = json.Marshal(trade_away)
	err = stub.PutState("_debug2", jsonAsBytes)
	
	
	open.Willing = append(open.Willing, trade_away)
	fmt.Println("! appended willing to open")
	jsonAsBytes, _ = json.Marshal(open)
	err = stub.PutState("_debug3", jsonAsBytes)
	
	
	trades.OpenTrades = append(trades.OpenTrades, open);						//append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(trades)
	err = stub.PutState("_opentrades", jsonAsBytes)								//rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("! open trade success ")
	return nil, nil
}

// ============================================================================================================================
// Perform Trade - close an open trade and move ownership
// ============================================================================================================================
func (t *SimpleChaincode) perform_trade(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//"bob", "444444444444", "asdf"
	if len(args) < 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}
	
	fmt.Println("! start close trade")
	timestamp, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("2nd argument must be a numeric string")
	}
	
	for i := range trades.OpenTrades{														//look for the trade
		fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found trade");
			
			t.set_user(stub, []string{args[0], args[2]})
			
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)	//remove trade
			jsonAsBytes, _ := json.Marshal(trades)
			err = stub.PutState("_opentrades", jsonAsBytes)									//rewrite open orders
			if err != nil {
				return nil, err
			}
		}
	}
	fmt.Println("! close trade success ")
	return nil, nil
}