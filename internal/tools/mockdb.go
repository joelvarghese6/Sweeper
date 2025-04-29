package tools

import (
	"time"
)

type mockDB struct{}

var mockLoginDetails = map[string]LoginDetails{
	"alex": {
		AuthToken: "123ABC",
		Username:  "alex",
	},
	"jaison": {
		AuthToken: "456DEF",
		Username:  "jaison",
	},
	"marie": {
		AuthToken: "789HGI",
		Username:  "marie",
	},
}	

var mockCoinDetails = map[string]CoinDetails{
	"alex": {
		Coins:     100,
		Username: "alex",
	},
	"jaison": {
		Coins:     200,
		Username: "jason",
	},
	"marie": {
		Coins:     300,
		Username: "marie",
	},
}

func (d *mockDB) GetUserLoginDetails(username string) *LoginDetails {
	//simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = LoginDetails{}
	clientData, ok := mockLoginDetails[username]

	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) GetUserCoins(username string) *CoinDetails {
	//simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = CoinDetails{}
	clientData, ok := mockCoinDetails[username]

	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) SetupDatabase() error {
	return nil
}