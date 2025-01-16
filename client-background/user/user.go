package user

import (
	"client-background/common"
	"client-background/types"
	"context"
	"encoding/json"
	"errors"
)

var descriptions = []string{"Success", "incorrect password", "account does not exist", "account already exists", "internal error"}

const InternalErrCode = 4

type Response struct {
	Status   int
	UserData types.UserData
}

func RegisterUser(userData types.UserData) (types.UserResponseParams, error) {
	funcName := "RegisterUser"
	endpoint := "register"
	var userRespParams types.UserResponseParams

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not marshal user data" + " : " + err.Error())
		return userRespParams, err
	}

	sessionDetails := types.SessionDetails{
		SessionID: "",
		Type:      "user",
	}

	ctx := context.WithValue(context.Background(), common.KEY, sessionDetails)

	resp, err := common.SendAndReceive(ctx, endpoint, userDataJSON)
	if err != nil {
		return userRespParams, err
	}

	var respStruct Response
	err = json.Unmarshal(resp, &respStruct)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not unmarshal response" + " : " + err.Error())
		return userRespParams, err
	}

	userRespParams.Code = respStruct.Status
	userRespParams.Description = descriptions[respStruct.Status]
	return userRespParams, nil
}

func VerifyUser(credentials types.UserData) (types.UserResponseParams, error) {
	funcName := "VerifyUser"
	endpoint := "verify"
	var userRespParams types.UserResponseParams

	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not marshal user data" + " : " + err.Error())
		return userRespParams, err
	}

	sessionDetails := types.SessionDetails{
		SessionID: "",
		Type:      "user",
	}

	ctx := context.WithValue(context.Background(), common.KEY, sessionDetails)

	resp, err := common.SendAndReceive(ctx, endpoint, credentialsJSON)
	if err != nil {
		return userRespParams, err
	}

	var respStruct Response
	err = json.Unmarshal(resp, &respStruct)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not unmarshal response" + " : " + err.Error())
		return userRespParams, err
	}

	userRespParams.Code = respStruct.Status
	userRespParams.Description = descriptions[respStruct.Status]
	return userRespParams, nil
}

func UpdateUser(userData types.UserData) (types.UserResponseParams, error) {
	funcName := "UpdateUser"
	endpoint := "update"
	var userRespParams types.UserResponseParams

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not marshal user data" + " : " + err.Error())
		return userRespParams, err
	}

	sessionDetails := types.SessionDetails{
		SessionID: "",
		Type:      "user",
	}

	ctx := context.WithValue(context.Background(), common.KEY, sessionDetails)

	resp, err := common.SendAndReceive(ctx, endpoint, userDataJSON)
	if err != nil {
		return userRespParams, err
	}

	var respStruct Response
	err = json.Unmarshal(resp, &respStruct)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not unmarshal response" + " : " + err.Error())
		return userRespParams, err
	}

	userRespParams.Code = respStruct.Status
	userRespParams.Description = descriptions[respStruct.Status]
	return userRespParams, nil
}

func LoginUser(credentials types.UserData) (types.UserResponseParams, error) {
	endpoint := "login"
	funcName := "LoginUser"
	var userRespParams types.UserResponseParams

	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not marshal user data" + " : " + err.Error())
		return userRespParams, err
	}

	sessionDetails := types.SessionDetails{
		SessionID: "",
		Type:      "user",
	}

	ctx := context.WithValue(context.Background(), common.KEY, sessionDetails)

	resp, err := common.SendAndReceive(ctx, endpoint, credentialsJSON)
	if err != nil {
		return userRespParams, err
	}

	var respStruct Response
	err = json.Unmarshal(resp, &respStruct)
	if err != nil {
		err = errors.New(funcName + " :: " + "Could not unmarshal response" + " : " + err.Error())
		return userRespParams, err
	}

	userRespParams.Code = respStruct.Status
	userRespParams.Description = descriptions[respStruct.Status]
	userRespParams.UserInfo = respStruct.UserData
	return userRespParams, nil
}
