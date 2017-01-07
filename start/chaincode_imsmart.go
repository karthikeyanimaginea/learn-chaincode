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
	PointsQuantity int  `json:"vptquantity"`
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
	Owners    []Owner `json:"owner"`
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


func (t *SimpleChaincode) createVendor(stub shim.ChaincodeStubInterface, vendorname string, vendortype string, points int) ([]byte, error) {
	fmt.Println("Creating account")

	// Obtain the username to associate with the account
	if len(args) != 1 {
		fmt.Println("Error obtaining username")
		return nil, errors.New("createAccount accepts a single username argument")
	}
	vendorname := vendorname
	vendortype := vendortype
	points := points

	// Build an account object for the user
	var assetIds []string
	suffix := "000v"
	prefix := vendorname + suffix

	var account = Vendor{ID: prefix, Name: vendorname, Type: vendortype, Points: point}

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

func GetVendor(ID string, stub shim.ChaincodeStubInterface) (Account, error) {
	var company Account
	companyBytes, err := stub.GetState(accountPrefix + ID)
	if err != nil {
		fmt.Println("Account not found " + ID)
		return company, errors.New("Account not found " + ID)
	}

	err = json.Unmarshal(companyBytes, &company)
	if err != nil {
		fmt.Println("Error unmarshalling account " + ID + "\n err:" + err.Error())
		return company, errors.New("Error unmarshalling account " + ID)
	}

	return company, nil
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: %s", err)
	}
}
