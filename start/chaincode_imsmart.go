package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var cpPrefix = "cp:"
var accountPrefix = "acct:"

type SimpleChaincode struct {
}

func generateCUSIPSuffix(issueDate string, days int) (string, error) {

	t, err := msToTime(issueDate)
	if err != nil {
		return "", err
	}

	maturityDate := t.AddDate(0, 0, days)
	month := int(maturityDate.Month())
	day := maturityDate.Day()

	suffix := seventhDigit[month] + eigthDigit[day]
	return suffix, nil

}

const (
	millisPerSecond = int64(time.Second / time.Millisecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
)

func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(msInt / millisPerSecond,
		(msInt % millisPerSecond) * nanosPerMillisecond), nil
}

type Regulator struct {
	CurrencyName string	      `json:"curname"`
	Points string             `json:"points"`
	Quantity int              `json:"quantity"`
}

type Vendor struct {
	ID string           `json:"id"`
	Name string	        `json:"vendorname"`
	Type string         `json:"vendortype"`
	PointsQuantity string  `json:"vptquantity"`
}

type Customer struct {
	ID string              `json:"id"`
	CustomerMobile string  `json:"customermobile"`
	Points int             `json:"points"`
}

type CP struct {
	CUSIP     string  `json:"cusip"`
	Ticker    string  `json:"ticker"`
	Par       float64 `json:"par"`
	Qty       int     `json:"qty"`
	Discount  float64 `json:"discount"`
	Maturity  int     `json:"maturity"`
	Owners    []Regulator `json:"regulator"`
	Issuer    string  `json:"issuer"`
	IssueDate string  `json:"issueDate"`
}

type Account struct {
	ID          string  `json:"id"`
	Prefix      string  `json:"prefix"`
	CashBalance float64 `json:"cashBalance"`
	AssetsIds   []string `json:"assetIds"`
}

type Transaction struct {
	CUSIP       string   `json:"cusip"`
	From        string   `json:"fromCompany"`
	To          string   `json:"toCompany"`
	Quantity    int      `json:"quantity"`
	Discount    float64  `json:"discount"`
}


func (t *SimpleChaincode) createVendor(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating account")

	// Obtain the username to associate with the account
//	if len(args) != 1 {
//		fmt.Println("Error obtaining username")
//		return nil, errors.New("createAccount accepts a single username argument")
//	}
  vendorID := args[0]
	vendorname := args[1]
	vendortype := args[2]
	points := args[3]

	// Build an account object for the user
//	suffix := "000v"
//	prefix := vendorname + suffix

	var account = Vendor{ID: vendorID, Name: vendorname, Type: vendortype, PointsQuantity: points}

	accountBytes, err := json.Marshal(&account)
	if err != nil {
		fmt.Println("error creating account" + account.ID)
		return nil, errors.New("Error creating account " + account.ID)
	}

	fmt.Println("Attempting to get state of any existing account for " + account.ID)
	existingBytes, err := stub.GetState(accountPrefix + account.ID)
	if err == nil {

		var company Vendor
		err = json.Unmarshal(existingBytes, &company)
		if err != nil {
			fmt.Println("Error unmarshalling account " + account.ID + "\n--->: " + err.Error())

			if strings.Contains(err.Error(), "unexpected end") {
				fmt.Println("No data means existing account found for " + account.ID + ", initializing account.")
				err = stub.PutState(accountPrefix + account.ID, accountBytes)

				if err == nil {
					fmt.Println("created account" + accountPrefix + account.ID)
					return nil, nil
				} else {
					fmt.Println("failed to create initialize account for " + account.ID)
					return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
				}
			} else {
				return nil, errors.New("Error unmarshalling existing account " + account.ID)
			}
		} else {
			fmt.Println("Account already exists for " + account.ID + " " + company.ID)
			return nil, errors.New("Can't reinitialize existing user " + account.ID)
		}
	} else {

		fmt.Println("No existing account found for " + account.ID + ", initializing account.")
		err = stub.PutState(accountPrefix + account.ID, accountBytes)

		if err == nil {
			fmt.Println("created account" + accountPrefix + account.ID)
			return nil, nil
		} else {
			fmt.Println("failed to create initialize account for " + account.ID)
			return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
		}

	}

}


func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Init firing. Function will be ignored: " + function)

	// Initialize the collection of commercial paper keys
	fmt.Println("Initializing smart keys collection")
	var blank []string
	blankBytes, _ := json.Marshal(&blank)
	err := stub.PutState("SmartKeys", blankBytes)
	if err != nil {
		fmt.Println("Failed to initialize smart key collection")
	}

	fmt.Println("Initialization complete")
	return nil, nil
}


func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke running. Function: " + function)

	if function == "createVendor" {
		return t.createVendor(stub, args)
	}


	return nil, errors.New("Received unknown function invocation: " + function)
}


func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Query running. Function: " + function)


	if function == "GetVendor" {
 	 return t.GetVendor(stub, args)
  }
   return nil, nil
}


func (t *SimpleChaincode) GetVendor(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var vendordet Vendor
  ID := args[0]
	//suffix := "000v"
	companyBytes, err := stub.GetState(ID)
	if err != nil {
		fmt.Println("Account not found " + ID)
		return nil, errors.New("Account not found " + ID)
	}

	err = json.Unmarshal(companyBytes, &vendordet)
	if err != nil {
		fmt.Println("Error unmarshalling account " + ID + "\n err:" + err.Error())
		return nil, errors.New("Error unmarshalling account " + ID)
	}

	return companyBytes, nil
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: %s", err)
	}
}


//lookup tables for last two digits of CUSIP
var seventhDigit = map[int]string{
	1:  "A",
	2:  "B",
	3:  "C",
	4:  "D",
	5:  "E",
	6:  "F",
	7:  "G",
	8:  "H",
	9:  "J",
	10: "K",
	11: "L",
	12: "M",
	13: "N",
	14: "P",
	15: "Q",
	16: "R",
	17: "S",
	18: "T",
	19: "U",
	20: "V",
	21: "W",
	22: "X",
	23: "Y",
	24: "Z",
}

var eigthDigit = map[int]string{
	1:  "1",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "A",
	11: "B",
	12: "C",
	13: "D",
	14: "E",
	15: "F",
	16: "G",
	17: "H",
	18: "J",
	19: "K",
	20: "L",
	21: "M",
	22: "N",
	23: "P",
	24: "Q",
	25: "R",
	26: "S",
	27: "T",
	28: "U",
	29: "V",
	30: "W",
	31: "X",
}
